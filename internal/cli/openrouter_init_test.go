package cli

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteEnvAssignmentCreatesAndOverwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.local")

	if err := writeEnvAssignment(path, projectSecretEnvName, "sk-or-v1-first", false); err != nil {
		t.Fatalf("write first assignment: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	if got := string(data); got != "OPENROUTER_API_KEY=sk-or-v1-first\n" {
		t.Fatalf("unexpected env file:\n%s", got)
	}

	if err := writeEnvAssignment(path, projectSecretEnvName, "sk-or-v1-second", false); err == nil {
		t.Fatalf("expected duplicate write without overwrite to fail")
	}

	if err := writeEnvAssignment(path, projectSecretEnvName, "sk-or-v1-second", true); err != nil {
		t.Fatalf("overwrite assignment: %v", err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read overwritten env file: %v", err)
	}
	if got := string(data); strings.Contains(got, "sk-or-v1-first") || !strings.Contains(got, "sk-or-v1-second") {
		t.Fatalf("overwrite did not replace secret:\n%s", got)
	}
}

func TestEnsureVarlockSchemaAddsOpenRouterEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.schema")
	if err := os.WriteFile(path, []byte("APP_ENV=development\n"), 0644); err != nil {
		t.Fatalf("seed schema: %v", err)
	}

	if err := ensureVarlockSchema(path); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		"APP_ENV=development",
		"# @sensitive @required @type=string(startsWith=sk-or-v1-)",
		"OPENROUTER_API_KEY=",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("schema missing %q:\n%s", want, got)
		}
	}
}

func TestWriteVarlockProjectSecretWritesOnlyEncryptedValue(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX shell fake varlock")
	}
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	fakeVarlock := filepath.Join(binDir, "varlock")
	script := "#!/bin/sh\n" +
		"test \"$1\" = encrypt || exit 2\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' 'varlock(\"local:test-cipher\")'\n"
	if err := os.WriteFile(fakeVarlock, []byte(script), 0755); err != nil {
		t.Fatalf("write fake varlock: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	opts := initProjectOptions{
		EnvFile:     filepath.Join(dir, ".env.local"),
		SchemaFile:  filepath.Join(dir, ".env.schema"),
		SecretsMode: secretsModeVarlock,
	}
	if err := writeVarlockProjectSecret(context.Background(), opts, "sk-or-v1-super-secret"); err != nil {
		t.Fatalf("write varlock secret: %v", err)
	}

	envData, err := os.ReadFile(opts.EnvFile)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	env := string(envData)
	if strings.Contains(env, "sk-or-v1-super-secret") {
		t.Fatalf("plaintext secret was written to env file:\n%s", env)
	}
	if !strings.Contains(env, `OPENROUTER_API_KEY=varlock("local:test-cipher")`) {
		t.Fatalf("encrypted varlock value missing:\n%s", env)
	}
}

func TestAutoModeRequiresVarlockWhenSchemaExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", filepath.Join(dir, "empty-bin"))
	if err := os.WriteFile(filepath.Join(dir, ".env.schema"), []byte("APP_ENV=development\n"), 0644); err != nil {
		t.Fatalf("seed schema: %v", err)
	}

	_, err := resolveInitSecretsMode(dir, secretsModeAuto)
	if err == nil {
		t.Fatalf("expected auto mode to fail when schema exists but varlock is unavailable")
	}
}
