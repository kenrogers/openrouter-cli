# OpenRouter CLI Customizations

This project uses the Speakeasy-generated Go/Cobra CLI as the base. The custom layer is intentionally small and focused on agent-safe authentication.

## Generated-vs-Custom Auth Seam

- Generated code owns endpoint command generation, config loading, credential resolution order, and SDK security injection.
- Custom OpenRouter code owns the user workflows that the generated CLI cannot infer from the OpenAPI spec: browser PKCE login, validation-before-store, OS credential storage policy, agent environment export/install, `doctor`, `exec`, project init, and `keys create-saved`.
- Generated command files may call small custom hooks, but new OpenRouter-specific auth behavior should live in `internal/cli/openrouter_*` files rather than endpoint packages.
- Prefer the generated SDK/client for authenticated API endpoint calls. Raw HTTP is reserved for bootstrapping flows where generated auth injection would be wrong, especially exchanging a PKCE code before a new credential exists.

## Auth and Agent UX

- `openrouter login` and `openrouter auth login` use a browser-based PKCE flow, exchange the authorization code for an API key, validate it with `/key`, and save it.
- `openrouter login --print-env --no-store` performs session-only PKCE auth and emits shell code for `OPENROUTER_API_KEY` without touching Keychain or writing plaintext.
- `openrouter env` emits shell code for exporting the active credential into the current shell.
- `openrouter env install` installs a shell startup hook so future local agents inherit `OPENROUTER_API_KEY`; `--plaintext` is available when the user explicitly wants a globally discoverable environment variable.
- `openrouter login` stores API keys only in the operating system credential store.
- Plaintext config-file fallback for API keys is disabled; plaintext shell-profile export requires explicit `openrouter env install --plaintext`.
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

This repository has generated files with intentional custom hook calls, and `gen.yaml` currently has `generation.persistentEdits: {}`. Treat regeneration as unsafe until the generated/custom auth seam above is verified after `speakeasy run`. If repeated regeneration becomes part of the workflow, enable and test Speakeasy persistent edits or keep these customizations as an explicit post-generation patch.
