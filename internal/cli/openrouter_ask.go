package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

func newOpenRouterAskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ask [prompt]",
		Short: "Ask a text model a question",
		Long: `Ask a text model through OpenRouter chat completions.

This is the low-friction path for agents and developers: pass a prompt, let the
CLI choose a current text model, and get a compact response without hand-built
JSON.`,
		Example: `  openrouter ask "write a TypeScript fetch example for OpenRouter"
  openrouter ask --model anthropic/claude-sonnet-4.6 --system "be brief" "explain PKCE"
  openrouter ask --file prompt.md --max-tokens 600 --json`,
		RunE: runOpenRouterAskCommand,
	}
	cmd.Flags().StringP("prompt", "p", "", "Prompt text (can also be provided as positional args)")
	cmd.Flags().StringP("file", "f", "", "Read prompt text from a file")
	cmd.Flags().Bool("stdin", false, "Read prompt text from stdin")
	cmd.Flags().StringP("model", "m", "auto", "Text model ID, or auto to choose a current text model")
	cmd.Flags().String("system", "", "Optional system message")
	cmd.Flags().Float64("temperature", -1, "Sampling temperature")
	cmd.Flags().Int64("max-tokens", 0, "Maximum tokens to generate")
	cmd.Flags().String("provider", "", "Provider routing JSON object")
	cmd.Flags().Bool("raw", false, "Print only the model's text response")
	return cmd
}

func runOpenRouterAskCommand(cmd *cobra.Command, args []string) error {
	prompt, err := textPromptFromFlags(cmd, args, "ask")
	if err != nil {
		return err
	}
	apiKey, keySource, err := openRouterAPIKey(cmd, true)
	if err != nil {
		return err
	}
	model, _ := cmd.Flags().GetString("model")
	model = strings.TrimSpace(model)
	if model == "" || strings.EqualFold(model, "auto") {
		model, err = selectWorkflowModel(cmd, "text", preferredTextModels, defaultTextModel)
		if err != nil {
			return err
		}
	}

	messages := []map[string]any{}
	if system, _ := cmd.Flags().GetString("system"); strings.TrimSpace(system) != "" {
		messages = append(messages, map[string]any{"role": "system", "content": strings.TrimSpace(system)})
	}
	messages = append(messages, map[string]any{"role": "user", "content": prompt})

	body := map[string]any{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}
	if cmd.Flags().Changed("temperature") {
		temperature, _ := cmd.Flags().GetFloat64("temperature")
		body["temperature"] = temperature
	}
	if cmd.Flags().Changed("max-tokens") {
		maxTokens, _ := cmd.Flags().GetInt64("max-tokens")
		body["max_tokens"] = maxTokens
	}
	if provider, _ := cmd.Flags().GetString("provider"); strings.TrimSpace(provider) != "" {
		var providerConfig map[string]any
		if err := json.Unmarshal([]byte(provider), &providerConfig); err != nil {
			return fmt.Errorf("invalid --provider JSON: %w", err)
		}
		body["provider"] = providerConfig
	}

	req, _, err := openRouterJSONRequest(cmd.Context(), cmd, http.MethodPost, "/chat/completions", body, apiKey)
	if err != nil {
		return err
	}
	var response map[string]any
	if _, err := openRouterDoJSON(cmd, req, &response, 2*time.Minute); err != nil {
		return err
	}
	if client.IsDryRun(cmd) {
		return nil
	}

	text := extractChatText(response)
	if text == "" {
		return output.AgentModeError(cmd,
			"empty_completion",
			"OpenRouter response did not include text content",
			[]string{"Run with `--debug` to inspect the raw response", "Try a different text model with `--model`"},
		)
	}
	raw, _ := cmd.Flags().GetBool("raw")
	if raw {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), text)
		return err
	}
	return output.Result(cmd, map[string]any{
		"model":      model,
		"text":       text,
		"prompt":     prompt,
		"key_source": keySource,
	})
}

func textPromptFromFlags(cmd *cobra.Command, args []string, commandName string) (string, error) {
	values := []string{}
	if flagPrompt, _ := cmd.Flags().GetString("prompt"); strings.TrimSpace(flagPrompt) != "" {
		values = append(values, strings.TrimSpace(flagPrompt))
	}
	if argPrompt := strings.TrimSpace(strings.Join(args, " ")); argPrompt != "" {
		values = append(values, argPrompt)
	}
	if filePath, _ := cmd.Flags().GetString("file"); strings.TrimSpace(filePath) != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read prompt file %s: %w", filePath, err)
		}
		values = append(values, strings.TrimSpace(string(data)))
	}
	if useStdin, _ := cmd.Flags().GetBool("stdin"); useStdin {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		values = append(values, strings.TrimSpace(string(data)))
	}
	nonEmpty := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			nonEmpty = append(nonEmpty, strings.TrimSpace(value))
		}
	}
	if len(nonEmpty) > 1 {
		return "", fmt.Errorf("provide the prompt through only one input source")
	}
	if len(nonEmpty) == 0 {
		return "", fmt.Errorf("missing prompt: run `openrouter %s \"your prompt\"`, use --prompt, --file, or --stdin", commandName)
	}
	return nonEmpty[0], nil
}

func extractChatText(response map[string]any) string {
	var parts []string
	choices, _ := response["choices"].([]any)
	for _, choice := range choices {
		choiceMap, _ := choice.(map[string]any)
		message, _ := choiceMap["message"].(map[string]any)
		if text := textFromChatContent(message["content"]); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
