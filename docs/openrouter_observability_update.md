## openrouter observability update

Update an observability destination

### Synopsis

Update an existing observability destination. Only the fields provided in the request body are updated. [Management key](/docs/guides/overview/auth/management-api-keys) required.

```
openrouter observability update [flags]
```

### Examples

```
  openrouter observability update --id 99999999-aaaa-bbbb-cccc-dddddddddddd
```

### Options

```
  -a, --api-key-hashes null   Optional allowlist of OpenRouter API key hashes. null clears the filter (all keys). Omitting leaves the current value. Must contain at least one hash if provided.
      --body string           Request body as JSON (alternative to individual flags). Can also be provided via stdin.
  -c, --config-param string   Provider-specific configuration fields to update. Masked values are ignored; unset fields keep their current value.
  -e, --enabled               Whether the destination is enabled.
  -f, --filter-rules string   JSON object
  -h, --help                  help for update
  -i, --id string             The destination ID (UUID). [required]
  -n, --name string           Human-readable name for the destination.
  -p, --privacy-mode          When true, request/response bodies are not forwarded — only metadata.
  -s, --sampling-rate float   Sampling rate between 0 and 1 (1 = 100%).
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

* [openrouter observability](openrouter_observability.md)	 - Observability endpoints
