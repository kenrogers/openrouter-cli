# OpenRouter Agent Setup Prompt

Use this when an agent lands on `openrouter.ai/agents` and needs local OpenRouter access.

```text
Install the OpenRouter CLI first, then use it for auth and project key setup.

1. Install:
   curl -fsSL https://raw.githubusercontent.com/kenrogers/openrouter-cli/main/scripts/install.sh | env OPENROUTER_VERSION=v0.1.4 bash
   export PATH="$HOME/.local/bin:/usr/local/bin:$PATH"
   hash -r 2>/dev/null || true
   openrouter version

   The version must be 0.1.4 or newer. If it is older, stop and report the install output, the output of `command -v openrouter`, and the output of `openrouter version`.

2. Authenticate:
   eval "$(openrouter login --print-env --no-store)"

   This opens OpenRouter in the browser, waits for the local callback, and exports OPENROUTER_API_KEY into the current shell without writing it to disk or Keychain. Do not ask the user to paste an API key into chat. If the browser cannot open automatically, run:
   eval "$(openrouter login --print-env --no-store --no-open)"
   Then show the printed URL to the user.

   Keep subsequent commands in the same shell session. If your agent runs each command in a fresh shell, run login and the project setup in one shell command, or use saved login when the OS credential store works and prefix later commands with `eval "$(openrouter env)" &&`.

   If login reports credential_store_unavailable or macOS shows a Keychain reset dialog, stop. Tell the user to cancel the dialog unless they intentionally want to reset their default keychain. For agent setup, prefer the session-only command with `--no-store`, which avoids Keychain entirely.

3. Verify:
   openrouter --json doctor
   openrouter --json whoami

4. Set up the current project:
   openrouter init --secrets auto

   If Varlock is available or the project already uses .env.schema, prefer:
   openrouter init --secrets varlock

   This creates a project-level OpenRouter key. Varlock mode stores only an encrypted resolver in the local env file and updates .env.schema so agents can see the required config without seeing the secret.

5. If a one-off tool command needs OPENROUTER_API_KEY without project env setup:
   openrouter exec -- <command>

Auth resolution order:
1. --api-key
2. OPENROUTER_API_KEY
3. OS credential store

If install or auth fails, stop and report the command output. Do not ask the user to paste an API key into chat. If auth is missing in an agent environment, run `eval "$(openrouter login --print-env --no-store)"` and let the browser flow guide the user. Never ask the user to paste an API key into chat unless they explicitly choose a one-shot command such as `openrouter login --key <key>`.
```
