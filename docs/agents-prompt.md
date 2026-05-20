# OpenRouter Agent Setup Prompt

Use this when an agent lands on `openrouter.ai/agents` and needs local OpenRouter access.

```text
Install the OpenRouter CLI first, then use it to make OPENROUTER_API_KEY available to local agents.

1. Install:
   curl -fsSL https://raw.githubusercontent.com/kenrogers/openrouter-cli/main/scripts/install.sh | env OPENROUTER_VERSION=v0.1.5 bash
   export PATH="$HOME/.local/bin:/usr/local/bin:$PATH"
   hash -r 2>/dev/null || true
   openrouter version

   The version must be 0.1.5 or newer. If it is older, stop and report the install output, the output of `command -v openrouter`, and the output of `openrouter version`.

2. Authenticate:
   eval "$(openrouter login --print-env --no-store)"

   This opens OpenRouter in the browser, waits for the local callback, and exports OPENROUTER_API_KEY into the current shell without writing it to disk or Keychain. Do not ask the user to paste an API key into chat. If the browser cannot open automatically, run:
   eval "$(openrouter login --print-env --no-store --no-open)"
   Then show the printed URL to the user.

   This makes OPENROUTER_API_KEY available in the current shell. Keep subsequent commands in the same shell session.

3. Make it available to future shell-launched agents:
   openrouter env install --plaintext

   This intentionally stores OPENROUTER_API_KEY in the user's shell startup file so future local agents can detect it from the environment. Do not print the key.

   If the user wants credential-store backed setup instead and their OS credential store works, they can run `openrouter login` followed by `openrouter env install`.

4. Verify:
   openrouter --json doctor
   openrouter --json whoami

5. If a one-off tool command needs OPENROUTER_API_KEY without project env setup:
   openrouter exec -- <command>

Auth resolution order:
1. --api-key
2. OPENROUTER_API_KEY
3. OS credential store

Do not run `openrouter init` unless the user specifically asks to create project-scoped keys and has a management API key. A normal OpenRouter API key is enough for building with OpenRouter.

If install or auth fails, stop and report the command output. Do not ask the user to paste an API key into chat. If auth is missing in an agent environment, run `eval "$(openrouter login --print-env --no-store)"` and let the browser flow guide the user. Never ask the user to paste an API key into chat unless they explicitly choose a one-shot command such as `openrouter login --key <key>`.
```
