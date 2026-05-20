## openrouter byok create

Create a BYOK provider credential

### Synopsis

Create a new bring-your-own-key (BYOK) provider credential. The raw key is encrypted at rest and never returned in API responses. Defaults to the authenticated entity's default workspace; use the `workspace_id` body field to scope to a different workspace. [Management key](/docs/guides/overview/auth/management-api-keys) required.

```
openrouter byok create [flags]
```

### Examples

```
  openrouter byok create --key sk-proj-abc123... --provider openai
```

### Options

```
      --allowed-models null     Optional allowlist of model slugs this credential may be used for. null means no restriction.
      --allowed-user-ids null   Optional allowlist of user IDs that may use this credential. null means no restriction.
      --body string             Request body as JSON (alternative to individual flags). Can also be provided via stdin.
      --disabled                Whether this credential should be created in a disabled state.
  -h, --help                    help for create
  -i, --is-fallback             Whether this credential is treated as a fallback — used only after non-fallback keys for the same provider have been tried.
  -k, --key string              The raw provider API key or credential. This value is encrypted at rest and never returned in API responses. [required]
  -n, --name string             Optional human-readable name for the credential.
  -p, --provider openai         The upstream provider this credential authenticates against, as a lowercase slug (e.g. openai, `anthropic`, `amazon-bedrock`). (options: ai21, aion-labs, akashml, alibaba, amazon-bedrock, amazon-nova, ambient, anthropic, arcee-ai, atlas-cloud, avian, azure, baidu, baseten, black-forest-labs, byteplus, cerebras, chutes, cirrascale, clarifai, cloudflare, cohere, crusoe, deepinfra, deepseek, dekallm, featherless, fireworks, friendli, gmicloud, google-ai-studio, google-vertex, groq, hyperbolic, inception, inceptron, inference-net, infermatic, inflection, io-net, ionstream, liquid, mancer, mara, minimax, mistral, modelrun, modular, moonshotai, morph, ncompass, nebius, nex-agi, nextbit, novita, nvidia, open-inference, openai, parasail, perceptron, perplexity, phala, poolside, recraft, reka, relace, sambanova, seed, siliconflow, sourceful, stepfun, streamlake, switchpoint, together, upstage, venice, wandb, xai, xiaomi, z-ai) [required]
  -w, --workspace-id string     Optional workspace ID. Defaults to the authenticated entity's default workspace.
```

### Options inherited from parent commands

```
      --agent-mode                      Enable structured errors and default TOON output for AI coding agents. Automatically enabled when a known agent environment is detected (CLAUDE_CODE, CURSOR_AGENT, etc.). Use --agent-mode=false to disable.
      --api-key string                  API key as bearer token in Authorization header
      --app-categories string           Comma-separated list of app categories (e (env: OPENROUTER_APP_CATEGORIES)
      --app-title string                The app display name allows you to customize how your app appears in OpenRouter's dashboard (env: OPENROUTER_APP_TITLE)
      --color string                    Control colored output: auto (color when output is a TTY), always, or never. Respects NO_COLOR and FORCE_COLOR env vars. (default "auto")
  -d, --debug                           Log request and response diagnostics to stderr
      --dry-run                         Preview the request that would be sent without executing it (output to stderr)
  -H, --header stringArray              Set a custom HTTP request header (format: "Key: Value"). Can be specified multiple times.
      --http-referer string             The app identifier should be your app's URL and is used as the primary identifier for rankings (env: OPENROUTER_HTTP_REFERER)
      --include-headers                 Include HTTP response headers in the output
  -q, --jq string                       Filter and transform output using a jq expression (e.g., '.name', '.items[] | .id')
      --json                            Shortcut for --output-format json
      --no-interactive                  Disable all interactive features (auto-prompting, explorer auto-launch, TUI forms)
      --no-retries                      Disable automatic retries (default: retries enabled with exponential backoff)
  -o, --output-format string            Specify the output format. Options: pretty, json, yaml, table, toon. (default "pretty")
      --retry-config string             Full retry config as JSON. Schema: {"strategy":"backoff","backoff":{"initialInterval":500,"maxInterval":10000,"exponent":1.5,"maxElapsedTime":30000},"retryConnectionErrors":false}. Times are in milliseconds.
      --retry-connection-errors         Retry on connection errors (EOF, reset, etc.)
      --retry-max-elapsed-time string   Maximum total time for retries (e.g., 30s, 5m). Default: 30s
      --server string                   Select a server by index (for indexed servers) or name (for named servers)
      --server-url string               Override the default server URL
      --timeout string                  HTTP request timeout (e.g., 30s, 5m, 100ms)
      --usage                           Print the CLI Usage schema in KDL format
```

### SEE ALSO

* [openrouter byok](openrouter_byok.md)	 - BYOK endpoints
