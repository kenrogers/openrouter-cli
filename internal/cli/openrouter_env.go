package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kenrogers/openrouter-cli/internal/config"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

const openRouterAPIKeyEnv = "OPENROUTER_API_KEY"

const (
	openRouterEnvBlockStart = "# >>> openrouter env >>>"
	openRouterEnvBlockEnd   = "# <<< openrouter env <<<"
)

func newOpenRouterEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Emit shell code for OPENROUTER_API_KEY",
		Long: `Emit shell code that exports OPENROUTER_API_KEY from the current
OpenRouter credential source.

Use this when an agent or local tool needs OpenRouter credentials in its
environment without writing the key to a shell profile:

  eval "$(openrouter env)"

If a saved OS credential is unavailable, this command can still re-export an
OPENROUTER_API_KEY that is already present in the current process environment.`,
		RunE: runOpenRouterEnvCommand,
	}
	cmd.Flags().String("shell", "auto", "Shell syntax to emit: auto, posix, fish, powershell, or cmd")
	cmd.Flags().Bool("quiet", false, "Suppress status messages on stderr")
	cmd.AddCommand(newOpenRouterEnvInstallCommand())
	cmd.AddCommand(newOpenRouterEnvUninstallCommand())
	return cmd
}

func runOpenRouterEnvCommand(cmd *cobra.Command, args []string) error {
	key, source := config.ResolveSecurityCredential(cmd, "api-key")
	key = normalizeAPIKey(key)
	if key == "" {
		return output.AgentModeError(cmd,
			"auth_error",
			"No OpenRouter API key found",
			[]string{
				"Run `eval \"$(openrouter login --print-env --no-store)\"` to authenticate and export OPENROUTER_API_KEY for this shell",
				"Or set OPENROUTER_API_KEY in the shell for this session",
			},
		)
	}
	if err := writeOpenRouterEnv(cmd.OutOrStdout(), cmd, key); err != nil {
		return err
	}
	quiet, _ := cmd.Flags().GetBool("quiet")
	if !quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "Exported %s from %s.\n", openRouterAPIKeyEnv, source)
	}
	return nil
}

func newOpenRouterEnvInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a shell startup hook for OPENROUTER_API_KEY",
		Long: `Install a managed shell startup block so future shells can expose
OPENROUTER_API_KEY to local agents and developer tools.

By default, this installs a secure loader that reads the saved OpenRouter
credential from the operating system credential store each time the shell
starts. To write the key itself into the shell startup file, pass --plaintext.
Plaintext mode is convenient for agent discovery, but any process or user that
can read the profile file can read the key.`,
		RunE: runOpenRouterEnvInstallCommand,
	}
	cmd.Flags().String("shell", "auto", "Shell profile syntax: auto, posix, fish, powershell, or cmd")
	cmd.Flags().String("profile-file", "", "Shell startup file to update (default: auto-detect)")
	cmd.Flags().Bool("plaintext", false, "Write the API key directly into the profile file")
	return cmd
}

func newOpenRouterEnvUninstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the OpenRouter shell startup hook",
		RunE:  runOpenRouterEnvUninstallCommand,
	}
	cmd.Flags().String("shell", "auto", "Shell profile syntax: auto, posix, fish, powershell, or cmd")
	cmd.Flags().String("profile-file", "", "Shell startup file to update (default: auto-detect)")
	return cmd
}

func runOpenRouterEnvInstallCommand(cmd *cobra.Command, args []string) error {
	shell, profilePath, err := envInstallTarget(cmd)
	if err != nil {
		return err
	}

	plaintext, _ := cmd.Flags().GetBool("plaintext")
	key := ""
	source := "keyring"
	if plaintext {
		key, source = config.ResolveSecurityCredential(cmd, "api-key")
		key = normalizeAPIKey(key)
		if key == "" {
			return output.AgentModeError(cmd,
				"auth_error",
				"No OpenRouter API key found",
				[]string{
					"Run `eval \"$(openrouter login --print-env --no-store)\"` first",
					"Or set OPENROUTER_API_KEY for this command",
				},
			)
		}
	} else if config.GetKeyringValue("api-key") == "" {
		return output.AgentModeError(cmd,
			"credential_store_unavailable",
			"No saved OpenRouter credential was found for the shell startup hook",
			[]string{
				"Run `openrouter login` to save the key in the OS credential store, then rerun `openrouter env install`",
				"Or run `eval \"$(openrouter login --print-env --no-store)\"` and then `openrouter env install --plaintext` to write the key into the profile file",
			},
		)
	}

	block, err := openRouterEnvInstallBlock(shell, plaintext, key)
	if err != nil {
		return err
	}
	if err := writeManagedEnvBlock(profilePath, block); err != nil {
		return err
	}

	if plaintext {
		fmt.Fprintf(cmd.ErrOrStderr(), "Installed plaintext %s from %s in %s.\n", openRouterAPIKeyEnv, source, profilePath)
		fmt.Fprintln(cmd.ErrOrStderr(), "The key is now stored directly in that profile file.")
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "Installed OpenRouter credential-store loader in %s.\n", profilePath)
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "Open a new shell, or source the profile file, for future agents to inherit OPENROUTER_API_KEY.")
	return nil
}

func installPlaintextOpenRouterEnv(cmd *cobra.Command, key string) (string, error) {
	shell, profilePath, err := envInstallTarget(cmd)
	if err != nil {
		return "", err
	}
	block, err := openRouterEnvInstallBlock(shell, true, normalizeAPIKey(key))
	if err != nil {
		return "", err
	}
	if err := writeManagedEnvBlock(profilePath, block); err != nil {
		return "", err
	}
	return profilePath, nil
}

func runOpenRouterEnvUninstallCommand(cmd *cobra.Command, args []string) error {
	_, profilePath, err := envInstallTarget(cmd)
	if err != nil {
		return err
	}
	removed, err := removeManagedEnvBlock(profilePath)
	if err != nil {
		return err
	}
	if removed {
		fmt.Fprintf(cmd.ErrOrStderr(), "Removed OpenRouter env hook from %s.\n", profilePath)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "No OpenRouter env hook found in %s.\n", profilePath)
	}
	return nil
}

func writeOpenRouterEnv(w io.Writer, cmd *cobra.Command, key string) error {
	shell, _ := cmd.Flags().GetString("shell")
	line, err := formatOpenRouterEnv(shell, key)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, line)
	return err
}

func formatOpenRouterEnv(shell, key string) (string, error) {
	syntax, err := normalizeShellSyntax(shell)
	if err != nil {
		return "", err
	}
	switch syntax {
	case "posix":
		return fmt.Sprintf("export %s=%s\n", openRouterAPIKeyEnv, quotePOSIXShell(key)), nil
	case "fish":
		return fmt.Sprintf("set -gx %s %s;\n", openRouterAPIKeyEnv, quotePOSIXShell(key)), nil
	case "powershell":
		return fmt.Sprintf("$env:%s = %s\n", openRouterAPIKeyEnv, quotePowerShell(key)), nil
	case "cmd":
		return fmt.Sprintf("set \"%s=%s\"\r\n", openRouterAPIKeyEnv, escapeCmdValue(key)), nil
	default:
		return "", fmt.Errorf("unsupported shell syntax %q", syntax)
	}
}

func envInstallTarget(cmd *cobra.Command) (string, string, error) {
	shellFlag, _ := cmd.Flags().GetString("shell")
	shell, err := normalizeShellSyntax(shellFlag)
	if err != nil {
		return "", "", err
	}
	if shell == "cmd" {
		return "", "", fmt.Errorf("env install does not support cmd profiles; use PowerShell or pass --profile-file with a POSIX/fish/PowerShell shell")
	}
	profilePath, _ := cmd.Flags().GetString("profile-file")
	if strings.TrimSpace(profilePath) == "" {
		profilePath, err = defaultEnvProfilePath(shell)
		if err != nil {
			return "", "", err
		}
	} else {
		profilePath = expandHomePath(profilePath)
	}
	return shell, filepath.Clean(profilePath), nil
}

func defaultEnvProfilePath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch shell {
	case "fish":
		return filepath.Join(home, ".config", "fish", "conf.d", "openrouter.fish"), nil
	case "powershell":
		if runtime.GOOS == "windows" {
			return filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), nil
		}
		return filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1"), nil
	default:
		sh := strings.ToLower(os.Getenv("SHELL"))
		if strings.Contains(sh, "zsh") {
			return filepath.Join(home, ".zshenv"), nil
		}
		if strings.Contains(sh, "bash") {
			return filepath.Join(home, ".bashrc"), nil
		}
		return filepath.Join(home, ".profile"), nil
	}
}

func openRouterEnvInstallBlock(shell string, plaintext bool, key string) (string, error) {
	var body string
	if plaintext {
		line, err := formatOpenRouterEnv(shell, key)
		if err != nil {
			return "", err
		}
		body = strings.TrimRight(line, "\r\n")
	} else {
		switch shell {
		case "posix":
			body = "if command -v openrouter >/dev/null 2>&1; then\n  eval \"$(openrouter env --quiet 2>/dev/null)\"\nfi"
		case "fish":
			body = "if command -q openrouter\n  openrouter env --quiet --shell fish 2>/dev/null | source\nend"
		case "powershell":
			body = "if (Get-Command openrouter -ErrorAction SilentlyContinue) {\n  Invoke-Expression (& openrouter env --quiet --shell powershell 2>$null)\n}"
		default:
			return "", fmt.Errorf("unsupported shell syntax %q", shell)
		}
	}
	return openRouterEnvBlockStart + "\n" + body + "\n" + openRouterEnvBlockEnd + "\n", nil
}

func writeManagedEnvBlock(path, block string) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read profile file: %w", err)
	}
	next := replaceManagedEnvBlock(string(data), block)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(next), 0600); err != nil {
		return fmt.Errorf("write profile file: %w", err)
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(path, 0600)
	}
	return nil
}

func replaceManagedEnvBlock(content, block string) string {
	start := strings.Index(content, openRouterEnvBlockStart)
	end := strings.Index(content, openRouterEnvBlockEnd)
	if start >= 0 && end >= start {
		end += len(openRouterEnvBlockEnd)
		for end < len(content) && (content[end] == '\n' || content[end] == '\r') {
			end++
		}
		return content[:start] + block + content[end:]
	}
	content = strings.TrimRight(content, "\r\n")
	if content != "" {
		content += "\n\n"
	}
	return content + block
}

func removeManagedEnvBlock(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read profile file: %w", err)
	}
	content := string(data)
	start := strings.Index(content, openRouterEnvBlockStart)
	end := strings.Index(content, openRouterEnvBlockEnd)
	if start < 0 || end < start {
		return false, nil
	}
	end += len(openRouterEnvBlockEnd)
	for end < len(content) && (content[end] == '\n' || content[end] == '\r') {
		end++
	}
	next := strings.TrimRight(content[:start], "\r\n")
	if suffix := strings.TrimLeft(content[end:], "\r\n"); suffix != "" {
		if next != "" {
			next += "\n\n"
		}
		next += suffix
	}
	if next != "" {
		next += "\n"
	}
	if err := os.WriteFile(path, []byte(next), 0600); err != nil {
		return false, fmt.Errorf("write profile file: %w", err)
	}
	return true, nil
}

func expandHomePath(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func normalizeShellSyntax(shell string) (string, error) {
	shell = strings.ToLower(strings.TrimSpace(shell))
	switch shell {
	case "", "auto":
		return autoShellSyntax(), nil
	case "sh", "bash", "zsh", "posix":
		return "posix", nil
	case "fish":
		return "fish", nil
	case "pwsh", "powershell":
		return "powershell", nil
	case "cmd", "cmd.exe":
		return "cmd", nil
	default:
		return "", fmt.Errorf("invalid --shell %q: expected auto, posix, fish, powershell, or cmd", shell)
	}
}

func autoShellSyntax() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	if strings.Contains(strings.ToLower(os.Getenv("SHELL")), "fish") {
		return "fish"
	}
	return "posix"
}

func quotePOSIXShell(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func quotePowerShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func escapeCmdValue(value string) string {
	replacer := strings.NewReplacer(
		"^", "^^",
		"&", "^&",
		"|", "^|",
		"<", "^<",
		">", "^>",
		"%", "%%",
	)
	return replacer.Replace(value)
}
