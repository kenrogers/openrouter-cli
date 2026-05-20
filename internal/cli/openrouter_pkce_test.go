package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestPKCES256Challenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got := pkceS256Challenge(verifier); got != want {
		t.Fatalf("challenge mismatch: got %q want %q", got, want)
	}
}

func TestBuildPKCEAuthURL(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("auth-url", openRouterAuthURL, "")

	got, err := buildPKCEAuthURL(cmd, "http://localhost:3000/openrouter-cli/callback", "challenge")
	if err != nil {
		t.Fatalf("build URL: %v", err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	if u.Scheme != "https" || u.Host != "openrouter.ai" || u.Path != "/auth" {
		t.Fatalf("unexpected auth URL base: %s", got)
	}
	q := u.Query()
	if q.Get("callback_url") != "http://localhost:3000/openrouter-cli/callback" {
		t.Fatalf("callback_url mismatch: %q", q.Get("callback_url"))
	}
	if q.Get("code_challenge") != "challenge" {
		t.Fatalf("code_challenge mismatch: %q", q.Get("code_challenge"))
	}
	if q.Get("code_challenge_method") != "S256" {
		t.Fatalf("code_challenge_method mismatch: %q", q.Get("code_challenge_method"))
	}
}

func TestPKCECallbackServerReceivesCode(t *testing.T) {
	port := freeLocalPort(t)
	resultCh, shutdown, err := startPKCECallbackServer("127.0.0.1", port)
	if err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer shutdown()

	res, err := http.Get("http://127.0.0.1:" + strconv.Itoa(port) + pkceCallbackPath + "?code=auth_code_123")
	if err != nil {
		t.Fatalf("get callback: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d body %s", res.StatusCode, string(body))
	}
	if !strings.Contains(string(body), "authorized") {
		t.Fatalf("callback page did not look successful: %s", string(body))
	}

	select {
	case result := <-resultCh:
		if result.Code != "auth_code_123" || result.Err != "" {
			t.Fatalf("unexpected callback result: %+v", result)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for callback result")
	}
}

func TestPKCECallbackServerReceivesCodeOnRoot(t *testing.T) {
	port := freeLocalPort(t)
	resultCh, shutdown, err := startPKCECallbackServer("127.0.0.1", port)
	if err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer shutdown()

	res, err := http.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/?code=auth_code_root")
	if err != nil {
		t.Fatalf("get callback: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d body %s", res.StatusCode, string(body))
	}

	select {
	case result := <-resultCh:
		if result.Code != "auth_code_root" || result.Err != "" {
			t.Fatalf("unexpected callback result: %+v", result)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for callback result")
	}
}

func TestExchangePKCECodeForKeyDoesNotSendExistingAPIKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "sk-or-v1-stale")

	var gotAuthorization string
	var gotBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/auth/keys" {
			t.Errorf("path = %s, want /auth/keys", r.URL.Path)
		}
		gotAuthorization = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"key":"sk-or-v1-new","user_id":"user_test"}`)
	}))
	defer server.Close()

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.Flags().String("server-url", "", "")
	cmd.Flags().String("api-key", "", "")
	if err := cmd.Flags().Set("server-url", server.URL); err != nil {
		t.Fatalf("set server-url: %v", err)
	}
	if err := cmd.Flags().Set("api-key", "sk-or-v1-stale"); err != nil {
		t.Fatalf("set api-key: %v", err)
	}

	key, userID, err := exchangePKCECodeForKey(cmd, "auth_code_123", "verifier_123")
	if err != nil {
		t.Fatalf("exchange code: %v", err)
	}
	if key != "sk-or-v1-new" {
		t.Fatalf("key = %q, want sk-or-v1-new", key)
	}
	if userID != "user_test" {
		t.Fatalf("userID = %q, want user_test", userID)
	}
	if gotAuthorization != "" {
		t.Fatalf("Authorization header = %q, want empty", gotAuthorization)
	}
	if gotBody["code"] != "auth_code_123" || gotBody["code_verifier"] != "verifier_123" || gotBody["code_challenge_method"] != "S256" {
		t.Fatalf("unexpected exchange body: %#v", gotBody)
	}
}

func freeLocalPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	defer listener.Close()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	n, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return n
}
