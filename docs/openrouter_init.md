# openrouter init

Provision a project-level OpenRouter API key and configure the current project.

```sh
openrouter init
```

By default, `init` uses Varlock when available. To require encrypted local env storage:

```sh
openrouter init --secrets varlock
```

Varlock mode:

- creates a new OpenRouter API key using the saved management credential
- pipes the key to `varlock encrypt`
- writes only `OPENROUTER_API_KEY=varlock("local:...")` to the local env file
- adds a sensitive `OPENROUTER_API_KEY` entry to `.env.schema`
- does not print the full key

Plaintext mode is available when a gitignored local env file is acceptable:

```sh
openrouter init --secrets plaintext
```

Useful flags:

```sh
--name <name>
--limit <usd>
--limit-reset daily|weekly|monthly
--workspace-id <id>
--env-file <path>
--schema-file <path>
--overwrite
```
