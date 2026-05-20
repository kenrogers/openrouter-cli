# OpenRouter CLI

Agent-friendly command-line access to OpenRouter.

This CLI is generated from `openapi.yaml` with Speakeasy, then lightly customized for the OpenRouter auth workflow agents need: secure login, key validation, keychain-backed storage, project key provisioning, Varlock-compatible env setup, and machine-readable output.

## Install locally

```sh
go install ./cmd/openrouter
```

Or build a local binary:

```sh
go build -o ./bin/openrouter ./cmd/openrouter
```

## Authenticate

Interactive login opens OpenRouter in the browser, completes a local PKCE callback, exchanges the authorization code for an API key, validates it with `/key`, and stores it in the operating system credential store.

```sh
openrouter login
openrouter whoami
openrouter doctor
```

Resolution order is:

1. `--api-key`
2. `OPENROUTER_API_KEY`
3. Saved keychain credential

API keys are not read from or written to plaintext config files. If secure credential storage is unavailable, `openrouter login` refuses to save the key and tells the user to use `OPENROUTER_API_KEY` for that invocation instead.

If a browser cannot be opened automatically, run `openrouter login --no-open` and open the printed URL yourself. Manual key input remains available as `openrouter login --key <key>` for break-glass cases, but agents should prefer the browser flow.

For one-off agent and local tool commands that require `OPENROUTER_API_KEY`, inject it into a child process without writing it to a shell profile:

```sh
openrouter exec -- your-agent-command
```

## Project setup

Most app projects should use a project-level OpenRouter key. Run this from the project root:

```sh
openrouter init
```

`init` creates a new OpenRouter API key using the saved management credential, writes it as `OPENROUTER_API_KEY`, and never prints the full secret.

Secret storage modes:

```sh
openrouter init --secrets auto
openrouter init --secrets varlock
openrouter init --secrets plaintext
```

`auto` uses Varlock when the `varlock` command is available. `varlock` mode pipes the new key to `varlock encrypt` and writes only an encrypted `varlock("local:...")` resolver to `.env.local` or `.env`; it also adds a sensitive `OPENROUTER_API_KEY` entry to `.env.schema` so agents can understand the project config without seeing the secret.

Use `plaintext` only when a gitignored local env file is acceptable. The CLI updates `.gitignore` for the target env file.

## API key management

Key management endpoints require an OpenRouter management key.

```sh
openrouter keys list
openrouter keys create-saved --name "agent-dev" --limit 10 --limit-reset monthly
openrouter keys update <hash> --limit 25
openrouter keys delete <hash>
```

Use `keys create-saved` when an agent should create a scoped key and immediately make it the active local credential. The new secret is stored in the OS credential store and only a masked value is printed.

## Agent-friendly output

The CLI emits human-readable output in a TTY and JSON in non-interactive contexts, CI, or when `--json` is passed.

```sh
openrouter --json whoami
openrouter --json doctor
openrouter --json models list
```

Errors use a stable JSON shape:

```json
{ "error": { "message": "No API key found", "code": "auth_error" } }
```

## Speakeasy generation

Speakeasy is configured in `.speakeasy/workflow.yaml` and `gen.yaml`.

Once authenticated with Speakeasy and enabled for CLI generation:

```sh
speakeasy auth login
speakeasy run --target openrouter-cli --skip-upload-spec --output console
```

The generated operation tree is the base of the project. OpenRouter-specific auth and agent UX live in small customizations on top of the generated CLI.

<!-- Start Summary [summary] -->
## Summary

OpenRouter API: OpenAI-compatible API with additional OpenRouter features

For more information about the API: [OpenRouter Documentation](https://openrouter.ai/docs)
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [OpenRouter CLI](#openrouter-cli)
  * [Install locally](#install-locally)
  * [Authenticate](#authenticate)
  * [Project setup](#project-setup)
  * [API key management](#api-key-management)
  * [Agent-friendly output](#agent-friendly-output)
  * [Speakeasy generation](#speakeasy-generation)
  * [CLI Installation](#cli-installation)
  * [Shell Completion](#shell-completion)
  * [CLI Example Usage](#cli-example-usage)
  * [Authentication](#authentication)
  * [Available Commands](#available-commands)
  * [Request Body Input](#request-body-input)
  * [Server Selection](#server-selection)
  * [Output Formats](#output-formats)
  * [Server-Sent Event Streaming](#server-sent-event-streaming)
  * [Pagination](#pagination)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Diagnostics](#diagnostics)

<!-- End Table of Contents [toc] -->

<!-- Start CLI Installation [installation] -->
## CLI Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/kenrogers/openrouter-cli/main/scripts/install.sh | bash
```

### Quick Install (Windows PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/kenrogers/openrouter-cli/main/scripts/install.ps1 | iex
```

### Go Install

Alternatively, install directly via Go:

```bash
go install github.com/kenrogers/openrouter-cli/cmd/openrouter@latest
```

### Manual Download

Download pre-built binaries for your platform from the [releases page](https://github.com/kenrogers/openrouter-cli/releases).
<!-- End CLI Installation [installation] -->

<!-- Start Shell Completion [completion] -->
## Shell Completion

Shell completions are available for Bash, Zsh, Fish, and PowerShell.

### Bash

```bash
# Add to ~/.bashrc:
source <(openrouter completion bash)

# Or install permanently:
openrouter completion bash > /etc/bash_completion.d/openrouter
```

### Zsh

```zsh
# Add to ~/.zshrc:
source <(openrouter completion zsh)

# Or install permanently:
openrouter completion zsh > "${fpath[1]}/_openrouter"
```

### Fish

```fish
openrouter completion fish | source

# Or install permanently:
openrouter completion fish > ~/.config/fish/completions/openrouter.fish
```

### PowerShell

```powershell
openrouter completion powershell | Out-String | Invoke-Expression
```
<!-- End Shell Completion [completion] -->

<!-- Start CLI Example Usage [usage] -->
## CLI Example Usage

### Example

```bash
openrouter analytics get-user-activity --api-key 'Bearer test_token' --date 2025-08-24 --api-key-hash abc123def456... --user-id user_abc123

```
<!-- End CLI Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

Authentication credentials are resolved in this order:

### 1. Command-line flags

Pass credentials directly as flags to any command:

```bash
openrouter --api-key <value> <command> [arguments]
```

### 2. Environment variables

Set credentials via environment variables:

| Variable | Description |
|----------|-------------|
| `OPENROUTER_API_KEY` | API key as bearer token in Authorization header |

### 3. OS Keychain (recommended for workstations)

Credentials are stored securely in your operating system's keychain when you run the browser-based login flow:

```bash
openrouter login
```

Secret credentials (tokens, API keys, passwords) are automatically stored in:
- **macOS**: Keychain
- **Linux**: GNOME Keyring / KWallet (via D-Bus Secret Service)
- **Windows**: Windows Credential Locker

If no keychain is available, `openrouter login` refuses to save the API key in plaintext. In CI, set `OPENROUTER_API_KEY` for the job instead.

### 4. Configuration file for non-secret settings

Run the interactive `configure` command to store non-secret settings:

```bash
openrouter configure
```

Configuration is stored in `~/.config/openrouter/config.yaml`.
<!-- End Authentication [security] -->

<!-- Start Available Commands [operations] -->
## Available Commands

<details open>
<summary>Available commands</summary>

### [analytics](docs/openrouter_analytics.md)

* [`get-user-activity`](docs/openrouter_analytics_get-user-activity.md) - Get user activity grouped by endpoint

### [tts](docs/openrouter_tts.md)

* [`create-speech`](docs/openrouter_tts_create-speech.md) - Create speech

### [stt](docs/openrouter_stt.md)

* [`create-transcription`](docs/openrouter_stt_create-transcription.md) - Create transcription

### [O-auth](docs/openrouter_O-auth.md)

* [`exchange-auth-code-for-API-key`](docs/openrouter_O-auth_exchange-auth-code-for-API-key.md) - Exchange authorization code for API key
* [`create-auth-code`](docs/openrouter_O-auth_create-auth-code.md) - Create authorization code

### [byok](docs/openrouter_byok.md)

* [`list`](docs/openrouter_byok_list.md) - List BYOK provider credentials
* [`create`](docs/openrouter_byok_create.md) - Create a BYOK provider credential
* [`delete`](docs/openrouter_byok_delete.md) - Delete a BYOK provider credential
* [`get`](docs/openrouter_byok_get.md) - Get a BYOK provider credential
* [`update`](docs/openrouter_byok_update.md) - Update a BYOK provider credential

### [chat](docs/openrouter_chat.md)

* [`send`](docs/openrouter_chat_send.md) - Create a chat completion

### [credits](docs/openrouter_credits.md)

* [`get-credits`](docs/openrouter_credits_get-credits.md) - Get remaining credits

### [embeddings](docs/openrouter_embeddings.md)

* [`generate`](docs/openrouter_embeddings_generate.md) - Submit an embedding request
* [`list-models`](docs/openrouter_embeddings_list-models.md) - List all embeddings models

### [endpoints](docs/openrouter_endpoints.md)

* [`list-zdr-endpoints`](docs/openrouter_endpoints_list-zdr-endpoints.md) - Preview the impact of ZDR on the available endpoints
* [`list`](docs/openrouter_endpoints_list.md) - List all endpoints for a model

### [generations](docs/openrouter_generations.md)

* [`get`](docs/openrouter_generations_get.md) - Get request & usage metadata for a generation
* [`list-generation-content`](docs/openrouter_generations_list-generation-content.md) - Get stored prompt and completion content for a generation

### [guardrails](docs/openrouter_guardrails.md)

* [`list`](docs/openrouter_guardrails_list.md) - List guardrails
* [`create`](docs/openrouter_guardrails_create.md) - Create a guardrail
* [`delete`](docs/openrouter_guardrails_delete.md) - Delete a guardrail
* [`get`](docs/openrouter_guardrails_get.md) - Get a guardrail
* [`update`](docs/openrouter_guardrails_update.md) - Update a guardrail
* [`list-guardrail-key-assignments`](docs/openrouter_guardrails_list-guardrail-key-assignments.md) - List key assignments for a guardrail
* [`bulk-assign-keys`](docs/openrouter_guardrails_bulk-assign-keys.md) - Bulk assign keys to a guardrail
* [`bulk-unassign-keys`](docs/openrouter_guardrails_bulk-unassign-keys.md) - Bulk unassign keys from a guardrail
* [`list-guardrail-member-assignments`](docs/openrouter_guardrails_list-guardrail-member-assignments.md) - List member assignments for a guardrail
* [`bulk-assign-members`](docs/openrouter_guardrails_bulk-assign-members.md) - Bulk assign members to a guardrail
* [`bulk-unassign-members`](docs/openrouter_guardrails_bulk-unassign-members.md) - Bulk unassign members from a guardrail
* [`list-key-assignments`](docs/openrouter_guardrails_list-key-assignments.md) - List all key assignments
* [`list-member-assignments`](docs/openrouter_guardrails_list-member-assignments.md) - List all member assignments

### [keys](docs/openrouter_keys.md)

* [`get-current-key-metadata`](docs/openrouter_keys_get-current-key-metadata.md) - Get current API key
* [`list`](docs/openrouter_keys_list.md) - List API keys
* [`create`](docs/openrouter_keys_create.md) - Create a new API key
* [`create-saved`](docs/openrouter_keys_create-saved.md) - Create an OpenRouter API key and save it securely
* [`delete`](docs/openrouter_keys_delete.md) - Delete an API key
* [`get`](docs/openrouter_keys_get.md) - Get a single API key
* [`update`](docs/openrouter_keys_update.md) - Update an API key

### [models](docs/openrouter_models.md)

* [`list`](docs/openrouter_models_list.md) - List all models and their properties
* [`count`](docs/openrouter_models_count.md) - Get total count of available models
* [`list-for-user`](docs/openrouter_models_list-for-user.md) - List models filtered by user provider preferences, privacy settings, and guardrails

### [observability](docs/openrouter_observability.md)

* [`list`](docs/openrouter_observability_list.md) - List observability destinations
* [`create`](docs/openrouter_observability_create.md) - Create an observability destination
* [`delete`](docs/openrouter_observability_delete.md) - Delete an observability destination
* [`get`](docs/openrouter_observability_get.md) - Get an observability destination
* [`update`](docs/openrouter_observability_update.md) - Update an observability destination

### [organization](docs/openrouter_organization.md)

* [`list-members`](docs/openrouter_organization_list-members.md) - List organization members

### [presets](docs/openrouter_presets.md)

* [`create-presets-chat-completions`](docs/openrouter_presets_create-presets-chat-completions.md) - Create a preset from a chat-completions request body

### [providers](docs/openrouter_providers.md)

* [`list`](docs/openrouter_providers_list.md) - List all providers

### [rerank](docs/openrouter_rerank.md)

* [`rerank`](docs/openrouter_rerank_rerank.md) - Submit a rerank request

### [beta](docs/openrouter_beta.md)

### [responses](docs/openrouter_beta_responses.md)

* [`send`](docs/openrouter_beta_responses_send.md) - Create a response

### [video-generation](docs/openrouter_video-generation.md)

* [`generate`](docs/openrouter_video-generation_generate.md) - Submit a video generation request
* [`get-generation`](docs/openrouter_video-generation_get-generation.md) - Poll video generation status
* [`get-video-content`](docs/openrouter_video-generation_get-video-content.md) - Download generated video content
* [`list-videos-models`](docs/openrouter_video-generation_list-videos-models.md) - List all video generation models

### [workspaces](docs/openrouter_workspaces.md)

* [`list`](docs/openrouter_workspaces_list.md) - List workspaces
* [`create`](docs/openrouter_workspaces_create.md) - Create a workspace
* [`delete`](docs/openrouter_workspaces_delete.md) - Delete a workspace
* [`get`](docs/openrouter_workspaces_get.md) - Get a workspace
* [`update`](docs/openrouter_workspaces_update.md) - Update a workspace
* [`bulk-add-members`](docs/openrouter_workspaces_bulk-add-members.md) - Bulk add members to a workspace
* [`bulk-remove-members`](docs/openrouter_workspaces_bulk-remove-members.md) - Bulk remove members from a workspace

</details>
<!-- End Available Commands [operations] -->

<!-- Start Request Body Input [stdinpiping] -->
## Request Body Input

Operations that accept a request body support three input methods, with a clear priority chain:

### Individual flags (highest priority)

```bash
openrouter <command> --name "Jane" --age 30
```

### `--body` flag

Provide the entire request body as a JSON string:

```bash
openrouter <command> --body '{"name": "John", "age": 30}'
```

Individual flags override `--body` values:

```bash
# Result: {name: "Jane", age: 30}
openrouter <command> --body '{"name": "John", "age": 30}' --name "Jane"
```

### Stdin piping (lowest priority)

Pipe JSON into any command that accepts a request body:

```bash
echo '{"name": "John", "age": 30}' | openrouter <command>
```

Individual flags override stdin values:

```bash
# Result: {name: "Jane", age: 30}
echo '{"name": "John", "age": 30}' | openrouter <command> --name "Jane"
```

This is useful for chaining commands, reading from files, or scripting:

```bash
# Read body from a file
openrouter <command> < request.json

# Pipe from another command
curl -s https://example.com/data.json | openrouter <command>
```

### Priority

When multiple input methods are used, the priority is:

| Priority | Source | Description |
|----------|--------|-------------|
| 1 (highest) | Individual flags | `--name "Jane"` always wins |
| 2 | `--body` flag | Whole-body JSON via flag |
| 3 (lowest) | Stdin | Piped JSON input |
<!-- End Request Body Input [stdinpiping] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Select Server by Name

Use `--server <name>` to select a named server (default: `production`):

| Name         | Server                         | Description       |
| ------------ | ------------------------------ | ----------------- |
| `production` | `https://openrouter.ai/api/v1` | Production server |

```bash
openrouter --server <name> <command> [arguments]
```

### Override Server URL

Use `--server-url` to override the server URL entirely, bypassing any named or indexed server selection:

```bash
openrouter --server-url https://custom-api.example.com <command> [arguments]
```

**Precedence**: `--server-url` > `--server` > default
<!-- End Server Selection [server] -->

<!-- Start Output Formats [output-formats] -->
## Output Formats

Every command supports a `--output-format` flag that controls how the response is rendered to stdout.

### Available formats

| Format | Flag | Description |
|--------|------|-------------|
| Pretty | `--output-format pretty` (default) | Aligned key-value pairs with color, nested indentation. Human-readable at a glance. |
| JSON | `--output-format json` | JSON output. Passthrough when the response is already JSON (preserves original field order and numeric precision). Falls back to typed marshaling otherwise. |
| YAML | `--output-format yaml` | YAML output via standard marshaling. |
| Table | `--output-format table` | Tabular output for array responses. |
| TOON | `--output-format toon` | [Token-Oriented Object Notation](https://github.com/toon-format/spec) — a compact, line-oriented format that typically uses 30–60% fewer tokens than JSON. Well-suited for piping responses into LLM prompts. |

```bash
# Default pretty output
openrouter <command>

# Machine-readable JSON
openrouter <command> --output-format json

# TOON for LLM-friendly compact output
openrouter <command> --output-format toon

# Pipe JSON to jq without using --output-format
openrouter <command> --output-format json | jq '.fieldName'
```

### jq filtering

Use `--jq` to filter or transform the response inline using a [jq](https://jqlang.org) expression. This always outputs JSON and overrides `--output-format`:

```bash
# Extract a single field
openrouter <command> --jq '.name'

# Filter an array
openrouter <command> --jq '.items[] | select(.active == true)'
```

### Color control

Use `--color` to control terminal colors:

| Value | Behavior |
|-------|----------|
| `auto` (default) | Color when stdout is a TTY, plain text otherwise |
| `always` | Always colorize |
| `never` | Never colorize |

The `NO_COLOR` and `FORCE_COLOR` environment variables are also respected.

### Streaming and pagination

When using `--all` (pagination) or streaming operations, output is written incrementally as items arrive:

| Format | Streaming behavior |
|--------|-------------------|
| `json` | One compact JSON object per line ([NDJSON](https://github.com/ndjson/ndjson-spec)) |
| `yaml` | YAML documents separated by `---` |
| `toon` | One TOON-encoded object per block, separated by blank lines |
| `pretty` (default) | Pretty-printed items separated by blank lines |
<!-- End Output Formats [output-formats] -->

<!-- Start Server-Sent Event Streaming [eventstreaming] -->
## Server-Sent Event Streaming

Some operations return server-sent events (SSE). These are streamed to the terminal in real-time, with each event output as a separate JSON object (one per line).

```bash
# Stream events in JSON format
openrouter --output-format json <streaming-command>

# Filter streaming events with jq
openrouter --output-format json <streaming-command> --jq '.data'
```

Events are output as they arrive. Use `Ctrl+C` to stop streaming.
<!-- End Server-Sent Event Streaming [eventstreaming] -->

<!-- Start Pagination [pagination] -->
## Pagination

Some operations in this CLI support automatic pagination. These operations accept `--all` to automatically fetch all pages and stream results incrementally.

### Basic usage

```bash
# Fetch a single page (default behavior)
openrouter <command> --page 1

# Automatically fetch all pages
openrouter <command> --all
```

### Limiting pages

Use `--max-pages` to cap the number of pages fetched:

```bash
# Fetch at most 5 pages
openrouter <command> --all --max-pages 5
```

### Output formats

When using `--all`, results are streamed as each page is fetched:

| Format | Behavior |
|--------|----------|
| `--output-format json` | One JSON object per line ([NDJSON](https://github.com/ndjson/ndjson-spec)) |
| `--output-format yaml` | YAML documents separated by `---` |
| `--output-format toon` | One TOON-encoded block per item, separated by blank lines |
| Default (pretty) | Pretty-printed items separated by blank lines |

```bash
# Stream all results as NDJSON
openrouter <command> --all --output-format json

# Pipe to jq for further processing
openrouter <command> --all --output-format json | jq '.fieldName'

# Use the built-in --jq flag
openrouter <command> --all --jq '.fieldName'
```

### How it works

Under the hood, `--all` calls the operation once, then follows the underlying `Next()` pagination closure to fetch subsequent pages. Results are written to stdout as they arrive rather than buffered in memory, so this works well even with large result sets.

Without `--all`, paginated operations behave like any other command — pass cursor, page, offset, or limit flags manually and get a single page of results.
<!-- End Pagination [pagination] -->

<!-- Start Retries [retries] -->
## Retries

Some operations in this CLI support automatic retries on failure. If the API returns a retryable status code (e.g., 503), the CLI will retry the request using an exponential backoff strategy.

### Disabling retries

```bash
# Disable all retries for a single command
openrouter <command> --no-retries
```

### Timeout

Set a per-request HTTP timeout:

```bash
openrouter <command> --timeout 30s
```

### Controlling retry behavior

```bash
# Cap the total time spent retrying
openrouter <command> --retry-max-elapsed-time 5s

# Retry on connection errors (EOF, connection reset, etc.)
openrouter <command> --retry-connection-errors
```

### Full retry configuration

For exact control over all retry parameters, pass a JSON configuration:

```bash
openrouter <command> --retry-config '{"strategy":"backoff","backoff":{"initialInterval":500,"maxInterval":60000,"exponent":1.5,"maxElapsedTime":300000}}'
```

### Persisting retry settings

Add retry configuration to your config file (`openrouter configure` or edit `~/.config/openrouter/config.yaml` directly):

```yaml
timeout: 30s
no_retries: false
retry_connection_errors: true
retry_max_elapsed_time: 1m
```

**Precedence**: `--no-retries` > `--retry-config` > individual flags > config file > API spec defaults
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

The CLI uses standard exit codes to indicate success or failure:

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success |
| `1` | Error (API error, invalid input, etc.) |

On success, the response data is printed to **stdout** as JSON. On failure, error details are printed to **stderr**.

```bash
# Capture output and handle errors
openrouter ... > output.json 2> error.log
if [ $? -ne 0 ]; then
  echo "Error occurred, see error.log"
fi
```
<!-- End Error Handling [errors] -->

<!-- Start Diagnostics [diagnostics] -->
## Diagnostics

The CLI includes two diagnostic flags available on all commands:

### Dry Run

Preview what would be sent without making any network calls:

```bash
openrouter <command> --dry-run
```

Output goes to stderr and includes:
- HTTP method and URL
- Request headers (sensitive values redacted)
- Request body preview (sensitive fields redacted)

The command exits successfully without contacting the API. This is useful for verifying request construction before executing.

### Debug

Log request and response diagnostics while running normally:

```bash
openrouter <command> --debug
```

Debug output goes to stderr and includes:
- Request method, URL, headers, and body preview
- Response status, headers, and body preview
- Transport errors (if any)

The command still executes normally and produces its regular output on stdout.

### Flag Precedence

If both `--dry-run` and `--debug` are set, `--dry-run` takes precedence and no network calls are made.

### Security

Sensitive information is automatically redacted in diagnostic output:
- **Headers**: `Authorization`, `Cookie`, `Set-Cookie`, `X-API-Key`, and other security headers show `[REDACTED]`
- **Body**: JSON fields named `password`, `secret`, `token`, `api_key`, `client_secret`, etc. show `[REDACTED]`

Diagnostic output should still be treated as potentially sensitive operational data.
<!-- End Diagnostics [diagnostics] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->
