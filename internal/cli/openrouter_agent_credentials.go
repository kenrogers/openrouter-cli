package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/kenrogers/openrouter-cli/internal/config"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

type agentCredentialExposureMode string

const (
	agentCredentialExposureModeSecure    agentCredentialExposureMode = "secure-loader"
	agentCredentialExposureModePlaintext agentCredentialExposureMode = "plaintext"
)

type agentCredentialRef struct {
	Key    string
	Source string
}

type agentCredentialExposureRequest struct {
	Mode       agentCredentialExposureMode
	Credential agentCredentialRef
}

type agentCredentialExposureResult struct {
	Mode             agentCredentialExposureMode
	Shell            string
	ProfilePath      string
	CredentialSource string
}

type agentCredentialDiagnostics struct {
	Key              string
	Source           string
	EnvPresent       bool
	EnvValue         string
	KeyringAvailable bool
	KeyringReason    string
	StartupHook      agentCredentialStartupHook
}

type agentCredentialStartupHook struct {
	ProfilePath string
	Installed   bool
	Mode        agentCredentialExposureMode
	Err         error
}

var getSavedOpenRouterAPIKey = func() string {
	return normalizeAPIKey(config.GetKeyringValue("api-key"))
}

func validateAgentCredentialLoginOptions(cmd *cobra.Command, noStore, printEnv, installEnv, plaintext bool) error {
	if plaintext && !installEnv {
		return output.AgentModeError(cmd,
			"invalid_auth_options",
			"`--plaintext` only applies when `--install-env` is used",
			[]string{
				"Run `openrouter login --install-env` to install the secure credential-store loader",
				"Or run `openrouter login --install-env --plaintext` if you intentionally want the key written to the shell profile",
			},
		)
	}
	if noStore && !printEnv && !installEnv {
		return output.AgentModeError(cmd,
			"invalid_auth_options",
			"`--no-store` requires `--print-env`, `--install-env --plaintext`, or both because the credential must go somewhere",
			[]string{
				"Run `openrouter login --install-env` to save the key securely and expose it to future shell-launched agents",
				"Run `openrouter login --print-env --no-store` for session-only auth",
				"Run `openrouter login --no-store --install-env --plaintext` only if plaintext profile storage is acceptable",
			},
		)
	}
	if noStore && installEnv && !plaintext {
		return output.AgentModeError(cmd,
			"invalid_auth_options",
			"`--no-store --install-env` cannot install the secure loader because no credential will be saved",
			[]string{
				"Run `openrouter login --install-env` to save the key in the OS credential store and install the secure loader",
				"Or run `openrouter login --no-store --install-env --plaintext` if plaintext profile storage is acceptable",
			},
		)
	}
	return nil
}

func agentCredentialExposureModeFromPlaintext(plaintext bool) agentCredentialExposureMode {
	if plaintext {
		return agentCredentialExposureModePlaintext
	}
	return agentCredentialExposureModeSecure
}

func resolveAgentCredential(cmd *cobra.Command) agentCredentialRef {
	key, source := config.ResolveSecurityCredential(cmd, "api-key")
	return agentCredentialRef{Key: normalizeAPIKey(key), Source: source}
}

func installAgentCredentialExposure(cmd *cobra.Command, req agentCredentialExposureRequest) (agentCredentialExposureResult, error) {
	mode := req.Mode
	if mode == "" {
		mode = agentCredentialExposureModeSecure
	}
	shell, profilePath, err := envInstallTarget(cmd)
	if err != nil {
		return agentCredentialExposureResult{}, err
	}

	var block string
	source := req.Credential.Source
	switch mode {
	case agentCredentialExposureModeSecure:
		if getSavedOpenRouterAPIKey() == "" {
			return agentCredentialExposureResult{}, output.AgentModeError(cmd,
				"credential_store_unavailable",
				"No saved OpenRouter credential was found for the secure shell startup hook",
				[]string{
					"Run `openrouter login --install-env` to save the key in the OS credential store and install the secure loader",
					"Or run `openrouter env install --plaintext` if plaintext profile storage is acceptable",
				},
			)
		}
		block, err = openRouterEnvInstallBlock(shell, false, "")
		source = "keyring"
	case agentCredentialExposureModePlaintext:
		credential := req.Credential
		credential.Key = normalizeAPIKey(credential.Key)
		if credential.Key == "" {
			credential = resolveAgentCredential(cmd)
		}
		if credential.Key == "" {
			return agentCredentialExposureResult{}, output.AgentModeError(cmd,
				"auth_error",
				"No OpenRouter API key found",
				[]string{
					"Run `openrouter login --install-env` to save the key securely and install the secure loader",
					"Or set OPENROUTER_API_KEY for this command before using plaintext mode",
				},
			)
		}
		block, err = openRouterEnvInstallBlock(shell, true, credential.Key)
		source = credential.Source
		if source == "" {
			source = "provided"
		}
	default:
		return agentCredentialExposureResult{}, fmt.Errorf("unsupported agent credential exposure mode %q", mode)
	}
	if err != nil {
		return agentCredentialExposureResult{}, err
	}
	if err := writeManagedEnvBlock(profilePath, block); err != nil {
		return agentCredentialExposureResult{}, err
	}
	return agentCredentialExposureResult{
		Mode:             mode,
		Shell:            shell,
		ProfilePath:      profilePath,
		CredentialSource: source,
	}, nil
}

func inspectAgentCredentialExposure(cmd *cobra.Command) agentCredentialDiagnostics {
	credential := resolveAgentCredential(cmd)
	envValue := normalizeAPIKey(os.Getenv(openRouterAPIKeyEnv))
	return agentCredentialDiagnostics{
		Key:              credential.Key,
		Source:           credential.Source,
		EnvPresent:       envValue != "",
		EnvValue:         envValue,
		KeyringAvailable: config.KeyringAvailable(),
		KeyringReason:    strings.TrimSpace(config.KeyringUnavailableReason()),
		StartupHook:      inspectDefaultAgentCredentialStartupHook(),
	}
}

func inspectDefaultAgentCredentialStartupHook() agentCredentialStartupHook {
	shell := autoShellSyntax()
	if shell == "cmd" {
		shell = "powershell"
	}
	profilePath, err := defaultEnvProfilePath(shell)
	if err != nil {
		return agentCredentialStartupHook{Err: err}
	}
	status := agentCredentialStartupHook{ProfilePath: profilePath}
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return status
		}
		status.Err = err
		return status
	}
	block, ok := managedEnvBlockContent(string(data))
	if !ok {
		return status
	}
	status.Installed = true
	status.Mode = classifyManagedEnvBlock(block)
	return status
}

func managedEnvBlockContent(content string) (string, bool) {
	start := strings.Index(content, openRouterEnvBlockStart)
	end := strings.Index(content, openRouterEnvBlockEnd)
	if start < 0 || end < start {
		return "", false
	}
	end += len(openRouterEnvBlockEnd)
	return content[start:end], true
}

func classifyManagedEnvBlock(block string) agentCredentialExposureMode {
	if strings.Contains(block, "openrouter env --quiet") {
		return agentCredentialExposureModeSecure
	}
	return agentCredentialExposureModePlaintext
}
