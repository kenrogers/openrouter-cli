package cli

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/kenrogers/openrouter-cli/internal/config"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

const openRouterAPIKeyEnv = "OPENROUTER_API_KEY"

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
	fmt.Fprintf(cmd.ErrOrStderr(), "Exported %s from %s.\n", openRouterAPIKeyEnv, source)
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
