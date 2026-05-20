# OpenRouter Agent Setup Prompt

Use this when an agent lands on `openrouter.ai/agents` and needs local OpenRouter access.

```text
Install the OpenRouter CLI first, then use it to make OPENROUTER_API_KEY available to local agents.

1. Install:
   curl -fsSL https://raw.githubusercontent.com/kenrogers/openrouter-cli/main/scripts/install.sh | env OPENROUTER_VERSION=v0.1.7 bash
   export PATH="$HOME/.local/bin:/usr/local/bin:$PATH"
   hash -r 2>/dev/null || true
   openrouter version

   The version must be 0.1.7 or newer. If it is older, stop and report the install output, the output of `command -v openrouter`, and the output of `openrouter version`.

2. Authenticate:
   eval "$(openrouter login --print-env --no-store --install-env)"

   This opens OpenRouter in the browser once, waits for the local callback, exports OPENROUTER_API_KEY into the current shell, and installs a managed shell startup block so future shell-launched agents can detect it. Do not ask the user to paste an API key into chat. If the browser cannot open automatically, run:
   eval "$(openrouter login --print-env --no-store --install-env --no-open)"
   Then show the printed URL to the user.

   This intentionally stores OPENROUTER_API_KEY in the user's shell startup file so future local agents can detect it from the environment. Do not print the key.

   If the user wants credential-store backed setup instead and their OS credential store works, they can run `openrouter login` followed by `openrouter env install`.

3. Verify:
   openrouter --json doctor
   openrouter --json whoami

4. If a one-off tool command needs OPENROUTER_API_KEY without project env setup:
   openrouter exec -- <command>

Auth resolution order:
1. --api-key
2. OPENROUTER_API_KEY
3. OS credential store

Do not run `openrouter init` unless the user specifically asks to create project-scoped keys and has a management API key. A normal OpenRouter API key is enough for building with OpenRouter.

If install or auth fails, stop and report the command output. Do not ask the user to paste an API key into chat. If auth is missing in an agent environment, run `eval "$(openrouter login --print-env --no-store --install-env)"` and let the browser flow guide the user. Never ask the user to paste an API key into chat unless they explicitly choose a one-shot command such as `openrouter login --key <key>`.
```
