package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateAgentCredentialLoginOptionsNoStoreInstallEnvRequiresPlaintext(t *testing.T) {
	cmd := &cobra.Command{}
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	err := validateAgentCredentialLoginOptions(cmd, true, false, true, false)
	if err == nil {
		t.Fatal("expected --no-store --install-env without --plaintext to fail")
	}
	if !strings.Contains(stderr.String(), "secure loader") {
		t.Fatalf("expected secure loader guidance, got:\n%s", stderr.String())
	}
}

func TestInstallAgentCredentialExposureSecureWritesLoaderWithoutPlaintext(t *testing.T) {
	oldGetSavedOpenRouterAPIKey := getSavedOpenRouterAPIKey
	getSavedOpenRouterAPIKey = func() string { return "sk-or-secure-test" }
	t.Cleanup(func() {
		getSavedOpenRouterAPIKey = oldGetSavedOpenRouterAPIKey
	})

	path := filepath.Join(t.TempDir(), ".zshenv")
	cmd := testAgentCredentialCommand(t, path)

	result, err := installAgentCredentialExposure(cmd, agentCredentialExposureRequest{
		Mode: agentCredentialExposureModeSecure,
	})
	if err != nil {
		t.Fatalf("install secure exposure: %v", err)
	}
	if result.Mode != agentCredentialExposureModeSecure || result.ProfilePath != path {
		t.Fatalf("unexpected exposure result: %+v", result)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, `openrouter env --quiet`) {
		t.Fatalf("secure profile block missing loader:\n%s", got)
	}
	if strings.Contains(got, "sk-or-secure-test") {
		t.Fatalf("secure profile block leaked plaintext key:\n%s", got)
	}
}

func TestInstallAgentCredentialExposurePlaintextWritesProvidedKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshenv")
	cmd := testAgentCredentialCommand(t, path)

	result, err := installAgentCredentialExposure(cmd, agentCredentialExposureRequest{
		Mode: agentCredentialExposureModePlaintext,
		Credential: agentCredentialRef{
			Key:    "sk-or-test'abc",
			Source: "pkce",
		},
	})
	if err != nil {
		t.Fatalf("install plaintext exposure: %v", err)
	}
	if result.Mode != agentCredentialExposureModePlaintext || result.CredentialSource != "pkce" {
		t.Fatalf("unexpected exposure result: %+v", result)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	want := "export OPENROUTER_API_KEY='sk-or-test'\"'\"'abc'"
	if !strings.Contains(string(data), want) {
		t.Fatalf("plaintext profile block missing quoted key:\n%s", string(data))
	}
}

func TestClassifyManagedEnvBlock(t *testing.T) {
	secure, err := openRouterEnvInstallBlock("posix", false, "")
	if err != nil {
		t.Fatalf("secure block: %v", err)
	}
	if got := classifyManagedEnvBlock(secure); got != agentCredentialExposureModeSecure {
		t.Fatalf("secure mode = %s", got)
	}

	plain, err := openRouterEnvInstallBlock("posix", true, "sk-or-test")
	if err != nil {
		t.Fatalf("plaintext block: %v", err)
	}
	if got := classifyManagedEnvBlock(plain); got != agentCredentialExposureModePlaintext {
		t.Fatalf("plaintext mode = %s", got)
	}
}

func testAgentCredentialCommand(t *testing.T, profilePath string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().String("shell", "posix", "")
	cmd.Flags().String("profile-file", "", "")
	if err := cmd.Flags().Set("profile-file", profilePath); err != nil {
		t.Fatalf("set profile path: %v", err)
	}
	return cmd
}
