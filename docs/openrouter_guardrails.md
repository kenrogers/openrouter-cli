## openrouter guardrails

Guardrails endpoints

### Synopsis

Guardrails endpoints

```
openrouter guardrails [flags]
```

### Options

```
  -h, --help   help for guardrails
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

* [openrouter](openrouter.md)	 - OpenRouter API: OpenAI-compatible API with additional OpenRouter features
* [openrouter guardrails bulk-assign-keys](openrouter_guardrails_bulk-assign-keys.md)	 - Bulk assign keys to a guardrail
* [openrouter guardrails bulk-assign-members](openrouter_guardrails_bulk-assign-members.md)	 - Bulk assign members to a guardrail
* [openrouter guardrails bulk-unassign-keys](openrouter_guardrails_bulk-unassign-keys.md)	 - Bulk unassign keys from a guardrail
* [openrouter guardrails bulk-unassign-members](openrouter_guardrails_bulk-unassign-members.md)	 - Bulk unassign members from a guardrail
* [openrouter guardrails create](openrouter_guardrails_create.md)	 - Create a guardrail
* [openrouter guardrails delete](openrouter_guardrails_delete.md)	 - Delete a guardrail
* [openrouter guardrails get](openrouter_guardrails_get.md)	 - Get a guardrail
* [openrouter guardrails list](openrouter_guardrails_list.md)	 - List guardrails
* [openrouter guardrails list-guardrail-key-assignments](openrouter_guardrails_list-guardrail-key-assignments.md)	 - List key assignments for a guardrail
* [openrouter guardrails list-guardrail-member-assignments](openrouter_guardrails_list-guardrail-member-assignments.md)	 - List member assignments for a guardrail
* [openrouter guardrails list-key-assignments](openrouter_guardrails_list-key-assignments.md)	 - List all key assignments
* [openrouter guardrails list-member-assignments](openrouter_guardrails_list-member-assignments.md)	 - List all member assignments
* [openrouter guardrails update](openrouter_guardrails_update.md)	 - Update a guardrail
