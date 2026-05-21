package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	osexec "os/exec"
	"runtime"
	"strings"
	"time"

	sdkclient "github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/config"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/kenrogers/openrouter-cli/internal/sdk/models/operations"
	"github.com/kenrogers/openrouter-cli/internal/sdk/optionalnullable"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	openRouterAuthURL        = "https://openrouter.ai/auth"
	openRouterDefaultAPIBase = "https://openrouter.ai/api/v1"
)

type keyMetadata struct {
	Hash            string   `json:"hash,omitempty"`
	Name            string   `json:"name,omitempty"`
	Label           string   `json:"label,omitempty"`
	Disabled        bool     `json:"disabled,omitempty"`
	Limit           *float64 `json:"limit,omitempty"`
	LimitRemaining  *float64 `json:"limit_remaining,omitempty"`
	LimitReset      *string  `json:"limit_reset,omitempty"`
	Usage           float64  `json:"usage,omitempty"`
	IsManagementKey bool     `json:"is_management_key,omitempty"`
	IsFreeTier      bool     `json:"is_free_tier,omitempty"`
	WorkspaceID     string   `json:"workspace_id,omitempty"`
}

type doctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type doctorReport struct {
	OK     bool          `json:"ok"`
	Checks []doctorCheck `json:"checks"`
}

type createSavedKeyResult struct {
	Saved       bool     `json:"saved"`
	SavedTo     string   `json:"saved_to"`
	KeyMasked   string   `json:"key_masked"`
	Hash        string   `json:"hash,omitempty"`
	Name        string   `json:"name,omitempty"`
	Limit       *float64 `json:"limit,omitempty"`
	LimitReset  *string  `json:"limit_reset,omitempty"`
	WorkspaceID string   `json:"workspace_id,omitempty"`
}

func newOpenRouterLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with OpenRouter",
		Long: `Authenticate with OpenRouter using a browser-based PKCE flow.

The CLI opens OpenRouter in your browser, waits for the local callback, exchanges
the authorization code for an API key, and stores it in the operating system
credential store. API keys are not written to config files.`,
		RunE: runAuthLoginCmd,
	}
	cmd.Flags().String("key", "", "API key to validate and store securely (manual fallback)")
	cmd.Flags().Bool("no-open", false, "Do not open the browser automatically; print the auth URL instead")
	cmd.Flags().Bool("print-env", false, "Print shell code that exports OPENROUTER_API_KEY for the current shell")
	cmd.Flags().Bool("no-store", false, "Do not save the API key; use with --print-env or --install-env --plaintext")
	cmd.Flags().Bool("install-env", false, "After login, install a managed shell startup block for future agents")
	cmd.Flags().Bool("plaintext", false, "With --install-env, write the API key directly into the shell profile instead of installing the secure loader")
	cmd.Flags().String("shell", "auto", "Shell syntax for --print-env: auto, posix, fish, powershell, or cmd")
	cmd.Flags().String("profile-file", "", "Shell startup file to update when --install-env is used (default: auto-detect)")
	cmd.Flags().String("auth-url", openRouterAuthURL, "OpenRouter browser authorization URL")
	cmd.Flags().String("callback-host", "localhost", "Loopback host for the PKCE callback")
	cmd.Flags().Int("callback-port", 3000, "Localhost port for the PKCE callback")
	cmd.Flags().Duration("login-timeout", 10*time.Minute, "How long to wait for browser authorization")
	return cmd
}

func newOpenRouterLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear saved OpenRouter credentials",
		Long:  "Clear saved OpenRouter credentials from the operating system credential store.",
		RunE:  runAuthLogoutCmd,
	}
}

func initOpenRouterUtilityCommands(parent *cobra.Command) {
	parent.AddCommand(newOpenRouterInitCommand())
	parent.AddCommand(newOpenRouterDoctorCommand())
	parent.AddCommand(newOpenRouterEnvCommand())
	parent.AddCommand(newOpenRouterExecCommand())
	initOpenRouterKeyCommands(parent)
}

func initOpenRouterKeyCommands(parent *cobra.Command) {
	keysCmd := findChildCommand(parent, "keys", "api-keys", "API-keys")
	if keysCmd == nil {
		return
	}
	keysCmd.AddCommand(newCreateSavedAPIKeyCommand())
}

func findChildCommand(parent *cobra.Command, names ...string) *cobra.Command {
	for _, child := range parent.Commands() {
		for _, name := range names {
			if child.Name() == name {
				return child
			}
			for _, alias := range child.Aliases {
				if alias == name {
					return child
				}
			}
		}
	}
	return nil
}

func runOpenRouterAuthLoginCmd(cmd *cobra.Command, args []string) error {
	key, source, err := loginKeyCandidate(cmd)
	if err != nil {
		return err
	}
	printEnv, _ := cmd.Flags().GetBool("print-env")
	noStore, _ := cmd.Flags().GetBool("no-store")
	installEnv, _ := cmd.Flags().GetBool("install-env")
	plaintextEnv, _ := cmd.Flags().GetBool("plaintext")
	if err := validateAgentCredentialLoginOptions(cmd, noStore, printEnv, installEnv, plaintextEnv); err != nil {
		return err
	}

	keyringAvailable := false
	if !noStore {
		keyringAvailable = config.KeyringAvailable()
	}
	if !noStore && !keyringAvailable && !printEnv && !installEnv {
		return credentialStoreUnavailableError(cmd)
	}

	if key == "" {
		var userID string
		key, userID, err = runOpenRouterPKCELogin(cmd)
		if err != nil {
			return err
		}
		source = "pkce"
		if userID != "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "Authorized OpenRouter user %s.\n", userID)
		}
	}

	key = normalizeAPIKey(key)
	if key == "" {
		return fmt.Errorf("no API key provided")
	}

	meta, err := validateOpenRouterAPIKey(cmd.Context(), cmd, key)
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	cfg := config.GetConfig()
	if cfg == nil {
		cfg = &config.Config{}
	}
	stored := false
	if keyringAvailable {
		if err := storeAPIKeySecurely(cfg, key); err != nil {
			return err
		}
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		stored = true
	}
	installedEnvProfile := ""
	installedEnvMode := agentCredentialExposureMode("")
	if installEnv {
		exposure, err := installAgentCredentialExposure(cmd, agentCredentialExposureRequest{
			Mode:       agentCredentialExposureModeFromPlaintext(plaintextEnv),
			Credential: agentCredentialRef{Key: key, Source: source},
		})
		if err != nil {
			return err
		}
		installedEnvProfile = exposure.ProfilePath
		installedEnvMode = exposure.Mode
	}

	out := cmd.OutOrStderr()
	fmt.Fprintf(out, "Authenticated with OpenRouter (%s).\n", source)
	if meta.Name != "" || meta.Label != "" || meta.Hash != "" {
		name := meta.Name
		if name == "" {
			name = meta.Label
		}
		if name == "" {
			name = meta.Hash
		}
		fmt.Fprintf(out, "Key: %s (%s)\n", name, maskSecret(key))
	}
	if stored {
		fmt.Fprintln(out, "Secret stored in the operating system credential store.")
		fmt.Fprintf(out, "Config saved to %s without storing the API key in plaintext.\n", config.GetConfigPath())
	} else if noStore && installedEnvProfile == "" {
		fmt.Fprintln(out, "Session-only auth: key was not saved.")
		fmt.Fprintln(out, "OPENROUTER_API_KEY was emitted for the current shell only.")
	} else if printEnv && installedEnvProfile == "" {
		fmt.Fprintln(out, "Secure credential storage is unavailable, so the key was not saved.")
		fmt.Fprintln(out, "OPENROUTER_API_KEY was emitted for the current shell only.")
	}
	if installedEnvProfile != "" {
		switch installedEnvMode {
		case agentCredentialExposureModeSecure:
			fmt.Fprintf(out, "Installed secure OPENROUTER_API_KEY loader in %s for future shell-launched agents.\n", installedEnvProfile)
		case agentCredentialExposureModePlaintext:
			if !stored {
				fmt.Fprintln(out, "Secret was not saved to the operating system credential store.")
			}
			fmt.Fprintf(out, "Installed plaintext OPENROUTER_API_KEY in %s for future shell-launched agents.\n", installedEnvProfile)
		default:
			fmt.Fprintf(out, "Installed OPENROUTER_API_KEY in %s for future shell-launched agents.\n", installedEnvProfile)
		}
	}
	if printEnv {
		if err := writeOpenRouterEnv(cmd.OutOrStdout(), cmd, key); err != nil {
			return err
		}
	}
	return nil
}

func credentialStoreUnavailableError(cmd *cobra.Command) error {
	reason := strings.TrimSpace(config.KeyringUnavailableReason())
	message := "Secure credential storage is unavailable; refusing to save the API key outside the OS credential store"
	if reason != "" {
		message = fmt.Sprintf("%s (%s)", message, reason)
	}
	return output.AgentModeError(cmd,
		"credential_store_unavailable",
		message,
		[]string{
			"On macOS, cancel any Keychain reset dialog unless you intentionally want to reset your default keychain",
			"Open Keychain Access and make sure the login keychain exists, is unlocked, and is the default keychain",
			"If the login keychain exists, run `security default-keychain -s ~/Library/Keychains/login.keychain-db` and retry `openrouter login`",
			"For one command only, set OPENROUTER_API_KEY in the shell instead of saving it globally",
		},
	)
}

func loginKeyCandidate(cmd *cobra.Command) (string, string, error) {
	if f := cmd.Flags().Lookup("key"); f != nil && f.Changed {
		return f.Value.String(), "--key", nil
	}
	if f := cmd.Flag("api-key"); f != nil && f.Changed {
		return f.Value.String(), "--api-key", nil
	}

	noInteractive, _ := cmd.Flags().GetBool("no-interactive")
	if noInteractive {
		if val := config.GetEnvValue("api-key"); val != "" {
			return val, "OPENROUTER_API_KEY", nil
		}
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		data, err := readSecret(cmd)
		if err == nil && strings.TrimSpace(string(data)) != "" {
			return string(data), "stdin", nil
		}
	}
	return "", "", nil
}

func normalizeAPIKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.TrimPrefix(key, "Bearer ")
	key = strings.TrimPrefix(key, "bearer ")
	return strings.TrimSpace(key)
}

func storeAPIKeySecurely(cfg *config.Config, key string) error {
	key = normalizeAPIKey(key)
	if key == "" {
		return fmt.Errorf("no API key provided")
	}
	if !config.KeyringAvailable() {
		return fmt.Errorf("secure credential storage is unavailable; refusing to write API key to plaintext config")
	}
	if err := config.SetKeyringValue("api-key", key); err != nil {
		return fmt.Errorf("secure credential storage is unavailable: %w", err)
	}
	cfg.Security.ApiKey = ""
	return nil
}

func validateOpenRouterAPIKey(ctx context.Context, cmd *cobra.Command, key string) (keyMetadata, error) {
	baseURL := openRouterDefaultAPIBase
	if f := cmd.Flag("server-url"); f != nil && f.Changed && strings.TrimSpace(f.Value.String()) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(f.Value.String()), "/")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/key", nil)
	if err != nil {
		return keyMetadata{}, err
	}
	req.Header.Set("Authorization", "Bearer "+normalizeAPIKey(key))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "openrouter-cli/"+Version)
	req.Header.Set("X-OpenRouter-Title", "OpenRouter CLI")
	req.Header.Set("HTTP-Referer", "https://openrouter.ai/agents")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return keyMetadata{}, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return keyMetadata{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return keyMetadata{}, fmt.Errorf("%s", openRouterErrorMessage(res.StatusCode, data))
	}

	var payload struct {
		Data keyMetadata `json:"data"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return keyMetadata{}, err
	}
	return payload.Data, nil
}

func openRouterErrorMessage(status int, data []byte) string {
	var payload struct {
		Error struct {
			Message string `json:"message"`
			Code    any    `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &payload); err == nil && payload.Error.Message != "" {
		if payload.Error.Code != nil {
			return fmt.Sprintf("%s (%v)", payload.Error.Message, payload.Error.Code)
		}
		return payload.Error.Message
	}
	body := strings.TrimSpace(string(data))
	if body == "" {
		return fmt.Sprintf("OpenRouter returned HTTP %d", status)
	}
	return fmt.Sprintf("OpenRouter returned HTTP %d: %s", status, body)
}

func openBrowser(url string) error {
	var cmd *osexec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = osexec.Command("open", url)
	case "windows":
		cmd = osexec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = osexec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func newOpenRouterDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check OpenRouter CLI authentication and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			report := doctorReport{OK: true}
			add := func(name, status, message string) {
				if status == "fail" {
					report.OK = false
				}
				report.Checks = append(report.Checks, doctorCheck{Name: name, Status: status, Message: message})
			}

			add("Config file", "pass", config.GetConfigPath())
			exposure := inspectAgentCredentialExposure(cmd)
			key := exposure.Key
			source := exposure.Source
			if path, err := osexec.LookPath("openrouter"); err == nil {
				add("PATH", "pass", path)
			} else {
				add("PATH", "warn", "openrouter was not found on PATH for this process")
			}
			if exposure.KeyringAvailable {
				add("Credential storage", "pass", "OS credential store is available")
			} else if key != "" && source == "env" {
				add("Credential storage", "warn", "OS credential store is unavailable; using OPENROUTER_API_KEY from the environment")
			} else if key != "" && source == "flag" {
				add("Credential storage", "warn", "OS credential store is unavailable; using --api-key for this invocation")
			} else {
				message := "OS credential store is unavailable; API keys will not be saved"
				if exposure.KeyringReason != "" {
					message += ": " + exposure.KeyringReason
				}
				add("Credential storage", "fail", message)
			}

			if exposure.EnvPresent {
				add("Agent environment", "pass", openRouterAPIKeyEnv+" is present in this process environment")
			} else if key != "" {
				add("Agent environment", "warn", fmt.Sprintf("Key is available from %s, but %s is not set for child agents", source, openRouterAPIKeyEnv))
			} else {
				add("Agent environment", "fail", openRouterAPIKeyEnv+" is not set")
			}
			if exposure.StartupHook.Err != nil {
				add("Agent startup hook", "warn", exposure.StartupHook.Err.Error())
			} else if exposure.StartupHook.Installed {
				add("Agent startup hook", "pass", fmt.Sprintf("%s installed in %s", exposure.StartupHook.Mode, exposure.StartupHook.ProfilePath))
			} else {
				add("Agent startup hook", "warn", fmt.Sprintf("No managed OpenRouter env hook found in %s", exposure.StartupHook.ProfilePath))
			}

			if key == "" {
				add("API key", "fail", "No API key found. Run `openrouter login`.")
				return output.Result(cmd, map[string]any{"ok": report.OK, "checks": report.Checks})
			}
			add("API key", "pass", fmt.Sprintf("%s from %s", maskSecret(key), source))

			meta, err := validateOpenRouterAPIKey(cmd.Context(), cmd, key)
			if err != nil {
				add("API validation", "fail", err.Error())
				return output.Result(cmd, map[string]any{"ok": report.OK, "checks": report.Checks})
			}
			label := meta.Name
			if label == "" {
				label = meta.Label
			}
			if label == "" {
				label = meta.Hash
			}
			add("API validation", "pass", fmt.Sprintf("OpenRouter accepted key %s", label))
			return output.Result(cmd, map[string]any{"ok": report.OK, "checks": report.Checks})
		},
	}
}

func newOpenRouterExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec -- <command> [args...]",
		Short: "Run a command with OPENROUTER_API_KEY injected from secure storage",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, source := config.ResolveSecurityCredential(cmd, "api-key")
			if key == "" {
				return output.AgentModeError(cmd,
					"auth_error",
					"No API key found",
					[]string{"Run `openrouter login` first", "Or set OPENROUTER_API_KEY for this invocation"},
				)
			}
			child := osexec.Command(args[0], args[1:]...)
			child.Env = append(os.Environ(), "OPENROUTER_API_KEY="+normalizeAPIKey(key))
			child.Stdin = cmd.InOrStdin()
			child.Stdout = cmd.OutOrStdout()
			child.Stderr = cmd.ErrOrStderr()
			if err := child.Run(); err != nil {
				return fmt.Errorf("command failed using API key from %s: %w", source, err)
			}
			return nil
		},
	}
}

func newCreateSavedAPIKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-saved",
		Short: "Create an OpenRouter API key and save it securely",
		Long: `Create an OpenRouter API key using the current authenticated key and
immediately save the new key to the operating system credential store.

The newly created key is shown only as a masked value.`,
		RunE: runCreateSavedAPIKeyCommand,
	}
	cmd.Flags().StringP("name", "n", "", "Name for the new API key")
	cmd.Flags().Float64("limit", 0, "Optional spending limit for the API key in USD")
	cmd.Flags().String("limit-reset", "", "Optional reset interval: daily, weekly, or monthly")
	cmd.Flags().String("workspace-id", "", "Optional workspace ID")
	cmd.Flags().Bool("include-byok-in-limit", false, "Include BYOK usage in the spending limit")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runCreateSavedAPIKeyCommand(cmd *cobra.Command, args []string) error {
	if key, _ := config.ResolveSecurityCredential(cmd, "api-key"); key == "" {
		return output.AgentModeError(cmd,
			"auth_error",
			"No management API key found",
			[]string{"Run `openrouter login` with a management key first", "Then rerun `openrouter keys create-saved --name <name>`"},
		)
	}

	name, _ := cmd.Flags().GetString("name")
	body := operations.CreateKeysRequestBody{Name: name}
	if cmd.Flags().Changed("limit") {
		limit, _ := cmd.Flags().GetFloat64("limit")
		body.Limit = optionalnullable.From(&limit)
	}
	if cmd.Flags().Changed("limit-reset") {
		value, _ := cmd.Flags().GetString("limit-reset")
		switch value {
		case "daily", "weekly", "monthly":
			reset := operations.LimitReset(value)
			body.LimitReset = optionalnullable.From(&reset)
		default:
			return fmt.Errorf("invalid --limit-reset %q: expected daily, weekly, or monthly", value)
		}
	}
	if cmd.Flags().Changed("workspace-id") {
		workspaceID, _ := cmd.Flags().GetString("workspace-id")
		body.WorkspaceID = &workspaceID
	}
	if cmd.Flags().Changed("include-byok-in-limit") {
		include, _ := cmd.Flags().GetBool("include-byok-in-limit")
		body.IncludeByokInLimit = &include
	}

	s, err := sdkclient.NewClient(cmd)
	if err != nil {
		return err
	}
	res, err := s.APIKeys.Create(cmd.Context(), operations.CreateKeysRequest{Body: body})
	if err != nil {
		return output.Error(cmd, err)
	}
	obj := res.GetObject()
	if obj == nil || obj.Key == "" {
		return fmt.Errorf("OpenRouter did not return a new API key to save")
	}

	cfg := config.GetConfig()
	if cfg == nil {
		cfg = &config.Config{}
	}
	if err := storeAPIKeySecurely(cfg, obj.Key); err != nil {
		return err
	}
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	data := obj.GetData()
	return output.Result(cmd, map[string]any{
		"saved":        true,
		"saved_to":     "OS credential store",
		"key_masked":   maskSecret(obj.Key),
		"hash":         data.Hash,
		"name":         data.Name,
		"limit":        data.Limit,
		"limit_reset":  data.LimitReset,
		"workspace_id": data.WorkspaceID,
	})
}
