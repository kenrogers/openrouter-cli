package cli

import (
	"io"
	"net"
	"net/http"
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
	resultCh, shutdown, err := startPKCECallbackServer(port)
	if err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer shutdown()

	res, err := http.Get("http://localhost:" + strconv.Itoa(port) + pkceCallbackPath + "?code=auth_code_123")
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

func freeLocalPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "localhost:0")
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
