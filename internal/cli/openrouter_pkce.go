package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/spf13/cobra"
)

const pkceCallbackPath = "/openrouter-cli/callback"

type pkceCallbackResult struct {
	Code string
	Err  string
}

func runOpenRouterPKCELogin(cmd *cobra.Command) (string, string, error) {
	host, _ := cmd.Flags().GetString("callback-host")
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}
	port, _ := cmd.Flags().GetInt("callback-port")
	timeout, _ := cmd.Flags().GetDuration("login-timeout")
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}

	verifier, err := generatePKCEVerifier()
	if err != nil {
		return "", "", err
	}
	challenge := pkceS256Challenge(verifier)
	callbackURL := (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, fmt.Sprintf("%d", port)),
		Path:   pkceCallbackPath,
	}).String()

	resultCh, shutdown, err := startPKCECallbackServer(host, port)
	if err != nil {
		return "", "", output.AgentModeError(cmd,
			"pkce_callback_unavailable",
			fmt.Sprintf("Could not start local PKCE callback server on %s: %v", net.JoinHostPort(host, fmt.Sprintf("%d", port)), err),
			[]string{
				"Stop the process using that port and rerun `openrouter login`",
				"Or rerun with `openrouter login --callback-port 3000` using a free OpenRouter-supported localhost port",
				"As a last resort, use `openrouter login --key <OPENROUTER_API_KEY>` outside chat",
			},
		)
	}
	defer shutdown()

	authURL, err := buildPKCEAuthURL(cmd, callbackURL, challenge)
	if err != nil {
		return "", "", err
	}

	noOpen, _ := cmd.Flags().GetBool("no-open")
	if noOpen {
		fmt.Fprintf(cmd.ErrOrStderr(), "Open this URL to authorize OpenRouter CLI:\n%s\n", authURL)
	} else if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Open this URL to authorize OpenRouter CLI:\n%s\n", authURL)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "Opened browser to authorize OpenRouter CLI.\nIf it did not open, use this URL:\n%s\n", authURL)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Waiting for OpenRouter browser authorization on %s ...\n", callbackURL)

	select {
	case result := <-resultCh:
		if result.Err != "" {
			return "", "", fmt.Errorf("OpenRouter authorization failed: %s", result.Err)
		}
		key, userID, err := exchangePKCECodeForKey(cmd, result.Code, verifier)
		if err != nil {
			return "", "", err
		}
		return key, userID, nil
	case <-time.After(timeout):
		return "", "", output.AgentModeError(cmd,
			"pkce_login_timeout",
			"Timed out waiting for OpenRouter browser authorization",
			[]string{
				"Rerun `openrouter login` and approve the browser prompt",
				"Use `openrouter login --no-open` to print the auth URL without opening a browser",
			},
		)
	case <-cmd.Context().Done():
		return "", "", cmd.Context().Err()
	}
}

func buildPKCEAuthURL(cmd *cobra.Command, callbackURL, challenge string) (string, error) {
	authURL, _ := cmd.Flags().GetString("auth-url")
	if strings.TrimSpace(authURL) == "" {
		authURL = openRouterAuthURL
	}
	u, err := url.Parse(authURL)
	if err != nil {
		return "", fmt.Errorf("invalid --auth-url: %w", err)
	}
	q := u.Query()
	q.Set("callback_url", callbackURL)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func startPKCECallbackServer(host string, port int) (<-chan pkceCallbackResult, func(), error) {
	if port <= 0 || port > 65535 {
		return nil, nil, fmt.Errorf("invalid callback port %d", port)
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, nil, err
	}

	resultCh := make(chan pkceCallbackResult, 1)
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}
	handleCallback := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		result := pkceCallbackResult{
			Code: strings.TrimSpace(query.Get("code")),
			Err:  strings.TrimSpace(firstNonEmpty(query.Get("error_description"), query.Get("error"))),
		}
		if result.Code == "" && result.Err == "" {
			result.Err = "missing authorization code"
		}
		select {
		case resultCh <- result:
		default:
		}
		writePKCECallbackPage(w, result)
	}
	mux.HandleFunc(pkceCallbackPath, handleCallback)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if strings.TrimSpace(query.Get("code")) != "" || strings.TrimSpace(query.Get("error")) != "" || strings.TrimSpace(query.Get("error_description")) != "" {
			handleCallback(w, r)
			return
		}
		http.NotFound(w, r)
	})

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			select {
			case resultCh <- pkceCallbackResult{Err: err.Error()}:
			default:
			}
		}
	}()

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}
	return resultCh, shutdown, nil
}

func writePKCECallbackPage(w http.ResponseWriter, result pkceCallbackResult) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if result.Err != "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "<!doctype html><title>OpenRouter CLI authorization failed</title><h1>Authorization failed</h1><p>%s</p><p>You can close this tab and return to your agent.</p>", html.EscapeString(result.Err))
		return
	}
	fmt.Fprint(w, "<!doctype html><title>OpenRouter CLI authorized</title><h1>OpenRouter CLI authorized</h1><p>You can close this tab and return to your agent.</p>")
}

func exchangePKCECodeForKey(cmd *cobra.Command, code, verifier string) (string, string, error) {
	baseURL := openRouterDefaultAPIBase
	if f := cmd.Flag("server-url"); f != nil && f.Changed && strings.TrimSpace(f.Value.String()) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(f.Value.String()), "/")
	}

	body, err := json.Marshal(map[string]string{
		"code":                  code,
		"code_verifier":         verifier,
		"code_challenge_method": "S256",
	})
	if err != nil {
		return "", "", err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth/keys", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "openrouter-cli/"+Version)
	req.Header.Set("X-OpenRouter-Title", "OpenRouter CLI")
	req.Header.Set("HTTP-Referer", "https://openrouter.ai/agents")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", "", fmt.Errorf("%s", openRouterErrorMessage(res.StatusCode, data))
	}

	var payload struct {
		Key    string `json:"key"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", "", err
	}
	if normalizeAPIKey(payload.Key) == "" {
		return "", "", fmt.Errorf("OpenRouter did not return an API key")
	}
	return normalizeAPIKey(payload.Key), payload.UserID, nil
}

func generatePKCEVerifier() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate PKCE verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func pkceS256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
