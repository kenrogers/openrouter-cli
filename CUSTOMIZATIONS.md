# OpenRouter CLI Customizations

This project uses the Speakeasy-generated Go/Cobra CLI as the base. The custom layer is intentionally small and focused on agent-safe authentication.

## Auth and Agent UX

- `openrouter login` and `openrouter auth login` use a browser-based PKCE flow, exchange the authorization code for an API key, validate it with `/key`, and save it.
- `openrouter login --print-env --no-store` performs session-only PKCE auth and emits shell code for `OPENROUTER_API_KEY` without touching Keychain or writing plaintext.
- `openrouter env` emits shell code for exporting the active credential into the current shell.
- API keys are stored only in the operating system credential store.
- Plaintext config-file fallback for API keys is disabled.
- `openrouter doctor` reports credential storage, auth source, and API validation status.
- `openrouter exec -- <command>` injects `OPENROUTER_API_KEY` into a child process without writing it to shell profiles or project files.
- `openrouter init` creates a project-level OpenRouter key and configures the current project with `OPENROUTER_API_KEY`.
- `openrouter init --secrets varlock` pipes the created key through `varlock encrypt`, writes only a `varlock("local:...")` resolver to the project env file, and updates `.env.schema` with a sensitive OpenRouter entry.
- `openrouter init --secrets plaintext` writes to a local env file and ensures that file is covered by `.gitignore`.
- `--json` is a shortcut for `--output-format json`.

## API Key Management

- The generated `API-keys` command is exposed as `keys`, while keeping `API-keys` and `api-keys` as aliases.
- `openrouter keys create-saved` creates a key through the generated SDK and immediately stores the returned secret in the OS credential store, printing only a masked key.

## Agent Prompt

- `docs/agents-prompt.md` contains the proposed prompt text for `openrouter.ai/agents`.

## Regeneration Note

This folder is not currently a Git repository, so Speakeasy persistent custom-code merging is not enabled yet. Before repeated regeneration, put the project under Git and enable `generation.persistentEdits.enabled` in `gen.yaml`, or keep these customizations isolated in a patch applied after `speakeasy run`.
