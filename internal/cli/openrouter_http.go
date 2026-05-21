package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/config"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

func openRouterAPIBase(cmd *cobra.Command) string {
	baseURL := openRouterDefaultAPIBase
	if f := cmd.Flag("server-url"); f != nil && f.Changed && strings.TrimSpace(f.Value.String()) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(f.Value.String()), "/")
	}
	return baseURL
}

func openRouterAPIKey(cmd *cobra.Command, required bool) (string, string, error) {
	key, source := config.ResolveSecurityCredential(cmd, "api-key")
	key = normalizeAPIKey(key)
	if key == "" && required && !client.IsDryRun(cmd) {
		return "", "", output.AgentModeError(cmd,
			"auth_error",
			"No OpenRouter API key found",
			[]string{
				"Run `openrouter login` to authenticate",
				"Or run `eval \"$(openrouter login --print-env --no-store --install-env)\"` to make OPENROUTER_API_KEY available to local agents",
			},
		)
	}
	return key, source, nil
}

func openRouterJSONRequest(ctx context.Context, cmd *cobra.Command, method, path string, body any, apiKey string) (*http.Request, []byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, openRouterAPIBase(cmd)+path, bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "openrouter-cli/"+Version)
	req.Header.Set("X-OpenRouter-Title", "OpenRouter CLI")
	req.Header.Set("HTTP-Referer", "https://openrouter.ai/agents")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	return req, data, nil
}

func openRouterDoJSON(cmd *cobra.Command, req *http.Request, out any, defaultTimeout time.Duration) ([]byte, error) {
	timeout := defaultTimeout
	if f := cmd.Flag("timeout"); f != nil && strings.TrimSpace(f.Value.String()) != "" {
		parsed, err := time.ParseDuration(strings.TrimSpace(f.Value.String()))
		if err != nil {
			return nil, fmt.Errorf("invalid --timeout value %q: %w", f.Value.String(), err)
		}
		timeout = parsed
	}
	httpClient := client.WrapClientForDiagnostics(cmd, &http.Client{Timeout: timeout})
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if client.IsDryRun(cmd) {
		return data, nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("%s", openRouterErrorMessage(res.StatusCode, data))
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return nil, err
		}
	}
	return data, nil
}
