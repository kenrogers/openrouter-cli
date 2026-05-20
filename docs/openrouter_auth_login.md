## openrouter auth login

Authenticate with OpenRouter

### Synopsis

Authenticate with OpenRouter using a browser-based PKCE flow.

The CLI opens OpenRouter in your browser, waits for the local callback, exchanges
the authorization code for an API key, and stores it in the operating system
credential store. API keys are not written to config files.

```
openrouter auth login [flags]
```

### Options

```
      --auth-url string          OpenRouter browser authorization URL (default "https://openrouter.ai/auth")
      --callback-port int        Localhost port for the PKCE callback (default 3000)
  -h, --help                     help for login
      --key string               API key to validate and store securely (manual fallback)
      --login-timeout duration   How long to wait for browser authorization (default 10m0s)
      --no-open                  Do not open the browser automatically; print the auth URL instead
      --no-store                 Do not save the API key; use with --print-env for session-only auth
      --print-env                Print shell code that exports OPENROUTER_API_KEY for the current shell
      --shell string             Shell syntax for --print-env: auto, posix, fish, powershell, or cmd (default "auto")
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

* [openrouter auth](openrouter_auth.md)	 - Manage authentication credentials
