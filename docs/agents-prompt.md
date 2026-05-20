# OpenRouter Agent Setup Prompt

Use this when an agent lands on `openrouter.ai/agents` and needs local OpenRouter access.

```text
Install the OpenRouter CLI first, then use it for auth and project key setup.

1. Install:
   curl -fsSL https://raw.githubusercontent.com/kenrogers/openrouter-cli/main/scripts/install.sh | bash

2. Authenticate:
   openrouter login

   This opens https://openrouter.ai/keys, asks the user to paste an API key with hidden input, validates it with OpenRouter, and saves it to the operating system credential store. Do not write API keys into prompts, project files, shell profiles, or agent memory.

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

If auth is missing in an agent environment, ask the user to run `openrouter login` in their terminal. Never ask the user to paste an API key into chat unless they explicitly choose a one-shot command such as `openrouter login --key <key>`.
```
