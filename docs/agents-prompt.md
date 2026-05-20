# OpenRouter Agent Setup Prompt

Use this when an agent lands on `openrouter.ai/agents` and needs local OpenRouter access.

```text
Install the OpenRouter CLI first, then use it for auth and key management.

1. Install:
   curl -fsSL https://raw.githubusercontent.com/kenrogers/openrouter-cli/main/scripts/install.sh | bash

2. Authenticate:
   openrouter login

   This opens https://openrouter.ai/keys, asks the user to paste an API key with hidden input, validates it with OpenRouter, and saves it to the operating system credential store. Do not write API keys into prompts, project files, shell profiles, or agent memory.

3. Verify:
   openrouter --json doctor
   openrouter --json whoami

4. Use OpenRouter from tools that expect OPENROUTER_API_KEY:
   openrouter exec -- <command>

5. When the user has a management key and wants a separate scoped local key:
   openrouter keys create-saved --name agent-dev --limit 10 --limit-reset monthly

   This creates a key through OpenRouter and immediately replaces the local saved credential with the new key. The CLI only prints a masked value.

Auth resolution order:
1. --api-key
2. OPENROUTER_API_KEY
3. OS credential store

If auth is missing in an agent environment, ask the user to run `openrouter login` in their terminal. Never ask the user to paste an API key into chat unless they explicitly choose a one-shot command such as `openrouter login --key <key>`.
```
