## openrouter guardrails create

Create a guardrail

### Synopsis

Create a new guardrail for the authenticated user. [Management key](/docs/guides/overview/auth/management-api-keys) required.

```
openrouter guardrails create [flags]
```

### Examples

```
  openrouter guardrails create --name My New Guardrail
```

### Options

```
      --allowed-models string            Array of model identifiers (slug or canonical_slug accepted)
      --allowed-providers string         List of allowed provider IDs
      --body string                      Request body as JSON (alternative to individual flags). Can also be provided via stdin.
      --content-filter-builtins string   Builtin content filters to apply. The "flag" action is only supported for "regex-prompt-injection"; PII slugs (email, phone, ssn, credit-card, ip-address, person-name, address) accept "block" or "redact" only.
      --content-filters string           Custom regex content filters to apply to request messages
      --description string               Description of the guardrail
      --enforce-zdr string               Deprecated. Use enforce_zdr_anthropic, enforce_zdr_openai, enforce_zdr_google, and enforce_zdr_other instead. When provided, its value is copied into any of those per-provider fields that are not explicitly specified on the request.
      --enforce-zdr-anthropic string     Whether to enforce zero data retention for Anthropic models. Falls back to enforce_zdr when not provided.
      --enforce-zdr-google string        Whether to enforce zero data retention for Google models. Falls back to enforce_zdr when not provided.
      --enforce-zdr-openai string        Whether to enforce zero data retention for OpenAI models. Falls back to enforce_zdr when not provided.
      --enforce-zdr-other string         Whether to enforce zero data retention for models that are not from Anthropic, OpenAI, or Google. Falls back to enforce_zdr when not provided.
  -h, --help                             help for create
      --ignored-models string            Array of model identifiers to exclude from routing (slug or canonical_slug accepted)
      --ignored-providers string         List of provider IDs to exclude from routing
  -l, --limit-usd string                 Spending limit in USD
  -n, --name string                      Name for the new guardrail [required]
  -r, --reset-interval string            Interval at which the limit resets (daily, weekly, monthly) (options: daily, weekly, monthly)
  -w, --workspace-id string              The workspace to create the guardrail in. Defaults to the default workspace if not provided.
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

* [openrouter guardrails](openrouter_guardrails.md)	 - Guardrails endpoints
