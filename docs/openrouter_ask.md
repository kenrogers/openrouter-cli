## openrouter ask

Ask a text model a question

### Synopsis

Ask a text model through OpenRouter chat completions.

This is the low-friction path for agents and developers: pass a prompt, let the
CLI choose a current text model, and get a compact response without hand-built
JSON.

```
openrouter ask [prompt] [flags]
```

### Examples

```
  openrouter ask "write a TypeScript fetch example for OpenRouter"
  openrouter ask --model anthropic/claude-sonnet-4.6 --system "be brief" "explain PKCE"
  openrouter ask --file prompt.md --max-tokens 600 --json
```

### Options

```
  -f, --file string         Read prompt text from a file
  -h, --help                help for ask
      --max-tokens int      Maximum tokens to generate
  -m, --model string        Text model ID, or auto to choose a current text model (default "auto")
  -p, --prompt string       Prompt text (can also be provided as positional args)
      --provider string     Provider routing JSON object
      --raw                 Print only the model's text response
      --stdin               Read prompt text from stdin
      --system string       Optional system message
      --temperature float   Sampling temperature (default -1)
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
