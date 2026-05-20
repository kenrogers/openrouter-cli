## openrouter workspaces update

Update a workspace

### Synopsis

Update an existing workspace by ID or slug. [Management key](/docs/guides/overview/auth/management-api-keys) required.

```
openrouter workspaces update [flags]
```

### Examples

```
  openrouter workspaces update --id production
```

### Options

```
      --body string                           Request body as JSON (alternative to individual flags). Can also be provided via stdin.
      --default-image-model string            Default image model for this workspace
      --default-provider-sort string          Default provider sort preference (price, throughput, latency, exacto)
      --default-text-model string             Default text model for this workspace
      --description string                    New description for the workspace
  -h, --help                                  help for update
      --id string                             The workspace ID (UUID) or slug [required]
      --io-logging-api-key-ids string         Optional array of API key IDs to filter I/O logging
      --io-logging-sampling-rate float        Sampling rate for I/O logging (0.0001-1)
      --is-data-discount-logging-enabled      Whether data discount logging is enabled
      --is-observability-broadcast-enabled    Whether broadcast is enabled
      --is-observability-io-logging-enabled   Whether private logging is enabled
  -n, --name string                           New name for the workspace
  -s, --slug string                           New URL-friendly slug (lowercase alphanumeric segments separated by single hyphens, no leading/trailing hyphens)
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

* [openrouter workspaces](openrouter_workspaces.md)	 - Workspaces endpoints
