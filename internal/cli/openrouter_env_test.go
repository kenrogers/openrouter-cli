package cli

import "testing"

func TestFormatOpenRouterEnvPOSIX(t *testing.T) {
	got, err := formatOpenRouterEnv("posix", "sk-or-test'abc")
	if err != nil {
		t.Fatalf("format env: %v", err)
	}
	want := "export OPENROUTER_API_KEY='sk-or-test'\"'\"'abc'\n"
	if got != want {
		t.Fatalf("env line mismatch:\ngot  %q\nwant %q", got, want)
	}
}

func TestFormatOpenRouterEnvPowerShell(t *testing.T) {
	got, err := formatOpenRouterEnv("powershell", "sk-or-test'abc")
	if err != nil {
		t.Fatalf("format env: %v", err)
	}
	want := "$env:OPENROUTER_API_KEY = 'sk-or-test''abc'\n"
	if got != want {
		t.Fatalf("env line mismatch:\ngot  %q\nwant %q", got, want)
	}
}

func TestNormalizeShellSyntaxRejectsUnknownShell(t *testing.T) {
	if _, err := normalizeShellSyntax("mystery"); err == nil {
		t.Fatal("expected unknown shell to fail")
	}
}
