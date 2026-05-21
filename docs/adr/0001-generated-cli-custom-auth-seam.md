# ADR 0001: Generated CLI Plumbing and Custom OpenRouter Auth UX

Status: Accepted

## Context

The CLI is generated from the OpenRouter OpenAPI spec with Speakeasy. That generated layer is valuable because it tracks endpoint shape, request and response models, command registration, config loading, credential resolution, and SDK client security wiring.

OpenRouter also needs auth workflows that are not expressible as simple endpoint commands. Local agents need a browser login, a safe place to store credentials, a way to expose `OPENROUTER_API_KEY` to child processes, and diagnostics that explain where credentials are coming from.

## Decision

Keep the generated auth/config/client plumbing, and keep the custom OpenRouter auth layer as a workflow layer on top of it.

The generated layer owns:

- endpoint command generation
- global flags
- config loading
- `flag > env > keyring` security credential resolution
- SDK client construction and bearer-token injection

The custom OpenRouter layer owns:

- browser PKCE login
- validation before storing an API key
- refusal to write API keys to plaintext config
- OS credential store integration
- `openrouter env`, `openrouter env install`, and `openrouter exec`
- `openrouter doctor`
- `openrouter keys create-saved`
- project initialization helpers

## Rules

Generated endpoint command packages should stay generated. New OpenRouter-specific auth behavior belongs in `internal/cli/openrouter_*` files.

Generated root/auth/configure files may call small custom hooks, but those hooks should be the narrow seam between generated command scaffolding and custom behavior.

Authenticated endpoint calls should use the generated SDK/client when possible. Raw HTTP is allowed for auth bootstrapping paths where generated security injection would be incorrect, such as exchanging a PKCE authorization code before the new API key exists.

## Consequences

The custom auth system has real advantages as a workflow layer, not as a replacement for generated credential plumbing. This keeps OpenAPI regeneration useful while preserving the agent-safe login and credential exposure behavior that generated commands cannot provide.
