package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"

	sdkclient "github.com/kenrogers/openrouter-cli/internal/client"
	"github.com/kenrogers/openrouter-cli/internal/config"
	"github.com/kenrogers/openrouter-cli/internal/output"
	"github.com/kenrogers/openrouter-cli/internal/sdk/models/operations"
	"github.com/kenrogers/openrouter-cli/internal/sdk/optionalnullable"
	"github.com/spf13/cobra"
)

const (
	projectSecretEnvName = "OPENROUTER_API_KEY"

	secretsModeAuto      = "auto"
	secretsModePlaintext = "plaintext"
	secretsModeVarlock   = "varlock"
)

type initProjectOptions struct {
	ProjectDir            string
	EnvFile               string
	SchemaFile            string
	SecretsMode           string
	Name                  string
	Limit                 float64
	LimitSet              bool
	LimitReset            string
	WorkspaceID           string
	IncludeByokInLimit    bool
	IncludeByokInLimitSet bool
	Overwrite             bool
}

type initProjectResult struct {
	OK                bool     `json:"ok"`
	DryRun            bool     `json:"dry_run,omitempty"`
	AlreadyConfigured bool     `json:"already_configured,omitempty"`
	KeyCreated        bool     `json:"key_created"`
	KeyMasked         string   `json:"key_masked,omitempty"`
	KeyHash           string   `json:"key_hash,omitempty"`
	KeyName           string   `json:"key_name,omitempty"`
	SecretsMode       string   `json:"secrets_mode"`
	EnvFile           string   `json:"env_file"`
	SchemaFile        string   `json:"schema_file,omitempty"`
	GitignoreFile     string   `json:"gitignore_file,omitempty"`
	GitignoreUpdated  bool     `json:"gitignore_updated,omitempty"`
	NextSteps         []string `json:"next_steps,omitempty"`
}

type provisionedProjectKey struct {
	Key         string
	Masked      string
	Hash        string
	Name        string
	WorkspaceID string
}

func (r initProjectResult) Map() map[string]any {
	out := map[string]any{
		"ok":           r.OK,
		"key_created":  r.KeyCreated,
		"secrets_mode": r.SecretsMode,
		"env_file":     r.EnvFile,
	}
	if r.DryRun {
		out["dry_run"] = true
	}
	if r.AlreadyConfigured {
		out["already_configured"] = true
	}
	if r.KeyMasked != "" {
		out["key_masked"] = r.KeyMasked
	}
	if r.KeyHash != "" {
		out["key_hash"] = r.KeyHash
	}
	if r.KeyName != "" {
		out["key_name"] = r.KeyName
	}
	if r.SchemaFile != "" {
		out["schema_file"] = r.SchemaFile
	}
	if r.GitignoreFile != "" {
		out["gitignore_file"] = r.GitignoreFile
	}
	if r.GitignoreUpdated {
		out["gitignore_updated"] = true
	}
	if len(r.NextSteps) > 0 {
		out["next_steps"] = r.NextSteps
	}
	return out
}

func newOpenRouterInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Provision OpenRouter credentials for the current project",
		Long: `Provision a project-level OpenRouter API key and write it to the
project environment in an agent-friendly way.

By default, the command uses Varlock when it is available and otherwise falls
back to a gitignored env file. Use --secrets varlock to require encrypted local
storage and avoid plaintext project env files.`,
		RunE: runOpenRouterInitCommand,
	}
	cmd.Flags().String("secrets", secretsModeAuto, "Secret storage mode: auto, varlock, or plaintext")
	cmd.Flags().String("project-dir", "", "Project directory to configure (default: auto-detect from current directory)")
	cmd.Flags().String("env-file", "", "Env file to update (default: .env.local for JS apps, otherwise .env)")
	cmd.Flags().String("schema-file", ".env.schema", "Varlock schema file to update when --secrets varlock is used")
	cmd.Flags().StringP("name", "n", "", "Name for the project API key (default: openrouter-<project-directory>)")
	cmd.Flags().Float64("limit", 0, "Optional spending limit for the project API key in USD")
	cmd.Flags().String("limit-reset", "", "Optional reset interval: daily, weekly, or monthly")
	cmd.Flags().String("workspace-id", "", "Optional workspace ID")
	cmd.Flags().Bool("include-byok-in-limit", false, "Include BYOK usage in the spending limit")
	cmd.Flags().Bool("overwrite", false, "Replace an existing OPENROUTER_API_KEY entry in the target env file")
	return cmd
}

func runOpenRouterInitCommand(cmd *cobra.Command, args []string) error {
	opts, err := readInitProjectOptions(cmd)
	if err != nil {
		return err
	}

	mode, err := resolveInitSecretsMode(opts.ProjectDir, opts.SecretsMode)
	if err != nil {
		return output.AgentModeError(cmd,
			"varlock_unavailable",
			err.Error(),
			[]string{
				"Install Varlock and rerun `openrouter init --secrets varlock`",
				"Or rerun with `--secrets plaintext` if a gitignored plaintext env file is acceptable",
			},
		)
	}
	opts.SecretsMode = mode

	if existing, err := envAssignmentExists(opts.EnvFile, projectSecretEnvName); err != nil {
		return err
	} else if existing && !opts.Overwrite {
		result := baseInitResult(opts)
		result.OK = true
		result.AlreadyConfigured = true
		result.NextSteps = initNextSteps(opts)
		return output.Result(cmd, result.Map())
	}

	if sdkclient.IsDryRun(cmd) {
		result := baseInitResult(opts)
		result.OK = true
		result.DryRun = true
		result.NextSteps = initNextSteps(opts)
		return output.Result(cmd, result.Map())
	}

	if key, _ := config.ResolveSecurityCredential(cmd, "api-key"); key == "" {
		return output.AgentModeError(cmd,
			"auth_error",
			"No management API key found",
			[]string{
				"Run `openrouter login` first",
				"Then rerun `openrouter init` from the project directory",
			},
		)
	}

	projectKey, err := createProjectAPIKey(cmd, opts)
	if err != nil {
		return err
	}

	result := baseInitResult(opts)
	result.KeyCreated = true
	result.KeyMasked = projectKey.Masked
	result.KeyHash = projectKey.Hash
	result.KeyName = projectKey.Name

	switch opts.SecretsMode {
	case secretsModeVarlock:
		if err := writeVarlockProjectSecret(cmd.Context(), opts, projectKey.Key); err != nil {
			return fmt.Errorf("created key %s but failed to write Varlock env file: %w", projectKey.Masked, err)
		}
		result.SchemaFile = opts.SchemaFile
	case secretsModePlaintext:
		if err := writeEnvAssignment(opts.EnvFile, projectSecretEnvName, projectKey.Key, opts.Overwrite); err != nil {
			return fmt.Errorf("created key %s but failed to write env file: %w", projectKey.Masked, err)
		}
	default:
		return fmt.Errorf("unsupported secrets mode %q", opts.SecretsMode)
	}

	gitignoreFile, updated, err := ensureEnvFileGitignored(opts.ProjectDir, opts.EnvFile)
	if err != nil {
		return err
	}
	result.OK = true
	result.GitignoreFile = gitignoreFile
	result.GitignoreUpdated = updated
	result.NextSteps = initNextSteps(opts)
	return output.Result(cmd, result.Map())
}

func readInitProjectOptions(cmd *cobra.Command) (initProjectOptions, error) {
	projectDirFlag, _ := cmd.Flags().GetString("project-dir")
	projectDir, err := resolveProjectDir(projectDirFlag)
	if err != nil {
		return initProjectOptions{}, err
	}

	envFileFlag, _ := cmd.Flags().GetString("env-file")
	envFile := envFileFlag
	if strings.TrimSpace(envFile) == "" {
		envFile = defaultEnvFile(projectDir)
	}
	envFile = resolveProjectPath(projectDir, envFile)

	schemaFileFlag, _ := cmd.Flags().GetString("schema-file")
	schemaFile := resolveProjectPath(projectDir, schemaFileFlag)

	name, _ := cmd.Flags().GetString("name")
	if strings.TrimSpace(name) == "" {
		name = "openrouter-" + filepath.Base(projectDir)
	}

	limit, _ := cmd.Flags().GetFloat64("limit")
	limitReset, _ := cmd.Flags().GetString("limit-reset")
	workspaceID, _ := cmd.Flags().GetString("workspace-id")
	includeByok, _ := cmd.Flags().GetBool("include-byok-in-limit")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	secretsMode, _ := cmd.Flags().GetString("secrets")

	return initProjectOptions{
		ProjectDir:            projectDir,
		EnvFile:               envFile,
		SchemaFile:            schemaFile,
		SecretsMode:           strings.ToLower(strings.TrimSpace(secretsMode)),
		Name:                  strings.TrimSpace(name),
		Limit:                 limit,
		LimitSet:              cmd.Flags().Changed("limit"),
		LimitReset:            strings.TrimSpace(limitReset),
		WorkspaceID:           strings.TrimSpace(workspaceID),
		IncludeByokInLimit:    includeByok,
		IncludeByokInLimitSet: cmd.Flags().Changed("include-byok-in-limit"),
		Overwrite:             overwrite,
	}, nil
}

func resolveProjectDir(flagValue string) (string, error) {
	if strings.TrimSpace(flagValue) != "" {
		return filepath.Abs(flagValue)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if root, ok := detectProjectRoot(cwd); ok {
		return root, nil
	}
	return cwd, nil
}

func detectProjectRoot(start string) (string, bool) {
	markers := []string{
		"package.json",
		"pyproject.toml",
		"go.mod",
		"Cargo.toml",
		"deno.json",
		"bun.lockb",
		"pnpm-lock.yaml",
		"yarn.lock",
		".git",
	}
	dir := filepath.Clean(start)
	for {
		for _, marker := range markers {
			if fileExists(filepath.Join(dir, marker)) {
				return dir, true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func defaultEnvFile(projectDir string) string {
	jsMarkers := []string{
		"package.json",
		"next.config.js",
		"next.config.mjs",
		"next.config.ts",
		"vite.config.js",
		"vite.config.mjs",
		"vite.config.ts",
	}
	for _, marker := range jsMarkers {
		if fileExists(filepath.Join(projectDir, marker)) {
			return ".env.local"
		}
	}
	return ".env"
}

func resolveProjectPath(projectDir, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(projectDir, path)
}

func resolveInitSecretsMode(projectDir, requested string) (string, error) {
	switch requested {
	case "", secretsModeAuto:
		if _, err := osexec.LookPath("varlock"); err == nil {
			return secretsModeVarlock, nil
		}
		if fileExists(filepath.Join(projectDir, ".env.schema")) {
			return "", errors.New("this project appears to use Varlock, but the `varlock` command is not installed")
		}
		return secretsModePlaintext, nil
	case secretsModeVarlock:
		if _, err := osexec.LookPath("varlock"); err != nil {
			return "", errors.New("Varlock is required for `--secrets varlock`, but the `varlock` command was not found")
		}
		return secretsModeVarlock, nil
	case secretsModePlaintext:
		return secretsModePlaintext, nil
	default:
		return "", fmt.Errorf("invalid --secrets %q: expected auto, varlock, or plaintext", requested)
	}
}

func createProjectAPIKey(cmd *cobra.Command, opts initProjectOptions) (provisionedProjectKey, error) {
	body := operations.CreateKeysRequestBody{Name: opts.Name}
	if opts.LimitSet {
		body.Limit = optionalnullable.From(&opts.Limit)
	}
	if opts.LimitReset != "" {
		switch opts.LimitReset {
		case "daily", "weekly", "monthly":
			reset := operations.LimitReset(opts.LimitReset)
			body.LimitReset = optionalnullable.From(&reset)
		default:
			return provisionedProjectKey{}, fmt.Errorf("invalid --limit-reset %q: expected daily, weekly, or monthly", opts.LimitReset)
		}
	}
	if opts.WorkspaceID != "" {
		body.WorkspaceID = &opts.WorkspaceID
	}
	if opts.IncludeByokInLimitSet {
		body.IncludeByokInLimit = &opts.IncludeByokInLimit
	}

	s, err := sdkclient.NewClient(cmd)
	if err != nil {
		return provisionedProjectKey{}, err
	}
	res, err := s.APIKeys.Create(cmd.Context(), operations.CreateKeysRequest{Body: body})
	if err != nil {
		return provisionedProjectKey{}, output.Error(cmd, err)
	}
	obj := res.GetObject()
	if obj == nil || obj.Key == "" {
		return provisionedProjectKey{}, fmt.Errorf("OpenRouter did not return a new API key")
	}
	data := obj.GetData()
	return provisionedProjectKey{
		Key:         normalizeAPIKey(obj.Key),
		Masked:      maskSecret(obj.Key),
		Hash:        data.Hash,
		Name:        data.Name,
		WorkspaceID: data.WorkspaceID,
	}, nil
}

func writeVarlockProjectSecret(ctx context.Context, opts initProjectOptions, key string) error {
	encrypted, err := encryptValueWithVarlock(ctx, key)
	if err != nil {
		return err
	}
	if err := ensureVarlockSchema(opts.SchemaFile); err != nil {
		return err
	}
	return writeEnvAssignment(opts.EnvFile, projectSecretEnvName, encrypted, opts.Overwrite)
}

func encryptValueWithVarlock(ctx context.Context, value string) (string, error) {
	cmd := osexec.CommandContext(ctx, "varlock", "encrypt")
	cmd.Stdin = strings.NewReader(normalizeAPIKey(value))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("varlock encrypt failed: %s", msg)
	}
	encrypted := parseVarlockEncryptOutput(string(out))
	if encrypted == "" {
		return "", fmt.Errorf("varlock encrypt did not return an encrypted value")
	}
	if !strings.HasPrefix(encrypted, "varlock(") {
		return "", fmt.Errorf("varlock encrypt returned an unexpected value")
	}
	return encrypted, nil
}

func parseVarlockEncryptOutput(out string) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if _, value, ok := strings.Cut(line, "="); ok {
			return strings.TrimSpace(value)
		}
		return line
	}
	return ""
}

func ensureVarlockSchema(path string) error {
	const block = `# OpenRouter API key for local development.
# @sensitive @required @type=string(startsWith=sk-or-v1-)
# @docs(https://openrouter.ai/docs)
OPENROUTER_API_KEY=
`
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read Varlock schema: %w", err)
	}
	if containsEnvAssignment(string(existing), projectSecretEnvName) {
		return nil
	}
	var next string
	if len(existing) > 0 {
		next = strings.TrimRight(string(existing), "\n") + "\n\n" + block
	} else {
		next = block
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(next), 0644)
}

func writeEnvAssignment(path, name, value string, overwrite bool) error {
	if value == "" {
		return fmt.Errorf("cannot write empty value for %s", name)
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read env file: %w", err)
	}
	lines := splitEnvLines(string(existing))
	replaced := false
	for i, line := range lines {
		if envLineName(line) == name {
			if !overwrite {
				return fmt.Errorf("%s already exists in %s; pass --overwrite to replace it", name, path)
			}
			lines[i] = name + "=" + value
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, name+"="+value)
	}
	next := strings.Join(lines, "\n")
	if next != "" {
		next += "\n"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(next), 0600); err != nil {
		return fmt.Errorf("write env file: %w", err)
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(path, 0600)
	}
	return nil
}

func envAssignmentExists(path, name string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read env file: %w", err)
	}
	return containsEnvAssignment(string(data), name), nil
}

func containsEnvAssignment(content, name string) bool {
	for _, line := range strings.Split(content, "\n") {
		if envLineName(line) == name {
			return true
		}
	}
	return false
}

func envLineName(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}
	line = strings.TrimPrefix(line, "export ")
	name, _, ok := strings.Cut(line, "=")
	if !ok {
		return ""
	}
	return strings.TrimSpace(name)
}

func splitEnvLines(content string) []string {
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}

func ensureEnvFileGitignored(projectDir, envFile string) (string, bool, error) {
	root := gitignoreRoot(projectDir)
	gitignorePath := filepath.Join(root, ".gitignore")
	rel, err := filepath.Rel(root, envFile)
	if err != nil || strings.HasPrefix(rel, "..") {
		rel = filepath.Base(envFile)
	}
	rel = filepath.ToSlash(rel)

	data, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return "", false, fmt.Errorf("read .gitignore: %w", err)
	}
	if gitignoreCovers(string(data), rel) {
		return gitignorePath, false, nil
	}
	next := strings.TrimRight(string(data), "\n")
	if next != "" {
		next += "\n"
	}
	next += rel + "\n"
	if err := os.WriteFile(gitignorePath, []byte(next), 0644); err != nil {
		return "", false, fmt.Errorf("write .gitignore: %w", err)
	}
	return gitignorePath, true, nil
}

func gitignoreRoot(projectDir string) string {
	dir := filepath.Clean(projectDir)
	for {
		if fileExists(filepath.Join(dir, ".git")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return projectDir
		}
		dir = parent
	}
}

func gitignoreCovers(content, rel string) bool {
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	base := filepath.Base(rel)
	for _, line := range strings.Split(content, "\n") {
		pattern := strings.TrimSpace(line)
		if pattern == "" || strings.HasPrefix(pattern, "#") || strings.HasPrefix(pattern, "!") {
			continue
		}
		pattern = filepath.ToSlash(strings.TrimPrefix(pattern, "./"))
		if pattern == rel || strings.TrimPrefix(pattern, "/") == rel {
			return true
		}
		if pattern == base || strings.TrimPrefix(pattern, "/") == base {
			return true
		}
		if ok, _ := filepath.Match(pattern, rel); ok {
			return true
		}
		if ok, _ := filepath.Match(pattern, base); ok {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func baseInitResult(opts initProjectOptions) initProjectResult {
	result := initProjectResult{
		SecretsMode: opts.SecretsMode,
		EnvFile:     opts.EnvFile,
	}
	if opts.SecretsMode == secretsModeVarlock {
		result.SchemaFile = opts.SchemaFile
	}
	return result
}

func initNextSteps(opts initProjectOptions) []string {
	switch opts.SecretsMode {
	case secretsModeVarlock:
		return []string{
			"Run project commands with `varlock run -- <command>` so OPENROUTER_API_KEY is injected at runtime",
			"Commit .env.schema if you want agents and teammates to see the required config shape",
		}
	default:
		return []string{
			fmt.Sprintf("Use %s through the project's normal env loading", projectSecretEnvName),
			fmt.Sprintf("Keep %s out of git", filepath.Base(opts.EnvFile)),
		}
	}
}
