## openrouter byok list

List BYOK provider credentials

### Synopsis

List the bring-your-own-key (BYOK) provider credentials for the authenticated entity's default workspace. Use the `workspace_id` query parameter to scope the result to a different workspace, or the `provider` query parameter to filter by upstream provider. [Management key](/docs/guides/overview/auth/management-api-keys) required.

```
openrouter byok list [flags]
```

### Examples

```
  openrouter byok list
```

### Options

```
  -a, --all                   Automatically paginate and fetch all results (streams NDJSON for JSON output)
  -h, --help                  help for list
  -l, --limit int             Maximum number of records to return (max 100)
      --max-pages int         Maximum number of pages to fetch when using --all (0 = no limit)
      --offset string         Number of records to skip for pagination
  -p, --provider openai       Optional provider slug to filter by (e.g. openai, `anthropic`, `amazon-bedrock`). (options: ai21, aion-labs, akashml, alibaba, amazon-bedrock, amazon-nova, ambient, anthropic, arcee-ai, atlas-cloud, avian, azure, baidu, baseten, black-forest-labs, byteplus, cerebras, chutes, cirrascale, clarifai, cloudflare, cohere, crusoe, deepinfra, deepseek, dekallm, featherless, fireworks, friendli, gmicloud, google-ai-studio, google-vertex, groq, hyperbolic, inception, inceptron, inference-net, infermatic, inflection, io-net, ionstream, liquid, mancer, mara, minimax, mistral, modelrun, modular, moonshotai, morph, ncompass, nebius, nex-agi, nextbit, novita, nvidia, open-inference, openai, parasail, perceptron, perplexity, phala, poolside, recraft, reka, relace, sambanova, seed, siliconflow, sourceful, stepfun, streamlake, switchpoint, together, upstage, venice, wandb, xai, xiaomi, z-ai)
  -w, --workspace-id string   Optional workspace ID to filter by. Defaults to the authenticated entity's default workspace.
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
