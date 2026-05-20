package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestOpenRouterEnvInstallBlockSecurePOSIX(t *testing.T) {
	got, err := openRouterEnvInstallBlock("posix", false, "")
	if err != nil {
		t.Fatalf("install block: %v", err)
	}
	for _, want := range []string{
		openRouterEnvBlockStart,
		`eval "$(openrouter env --quiet 2>/dev/null)"`,
		openRouterEnvBlockEnd,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("block missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "sk-or") {
		t.Fatalf("secure block should not contain a plaintext key:\n%s", got)
	}
}

func TestOpenRouterEnvInstallBlockPlaintextPOSIX(t *testing.T) {
	got, err := openRouterEnvInstallBlock("posix", true, "sk-or-test'abc")
	if err != nil {
		t.Fatalf("install block: %v", err)
	}
	want := "export OPENROUTER_API_KEY='sk-or-test'\"'\"'abc'"
	if !strings.Contains(got, want) {
		t.Fatalf("plaintext block missing quoted export:\n%s", got)
	}
}

func TestWriteManagedEnvBlockReplacesExistingBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshenv")
	original := "before\n\n" + openRouterEnvBlockStart + "\nold\n" + openRouterEnvBlockEnd + "\n\nafter\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	block := openRouterEnvBlockStart + "\nnew\n" + openRouterEnvBlockEnd + "\n"
	if err := writeManagedEnvBlock(path, block); err != nil {
		t.Fatalf("write block: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	got := string(data)
	if strings.Contains(got, "old") || !strings.Contains(got, "new") || !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Fatalf("managed block replacement failed:\n%s", got)
	}
}

func TestRemoveManagedEnvBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshenv")
	content := "before\n\n" + openRouterEnvBlockStart + "\nmanaged\n" + openRouterEnvBlockEnd + "\n\nafter\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	removed, err := removeManagedEnvBlock(path)
	if err != nil {
		t.Fatalf("remove block: %v", err)
	}
	if !removed {
		t.Fatal("expected block to be removed")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	got := string(data)
	if strings.Contains(got, "managed") || !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Fatalf("managed block removal failed:\n%s", got)
	}
}
