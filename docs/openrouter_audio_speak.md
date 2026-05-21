## openrouter audio speak

Generate speech audio from text

### Synopsis

Generate speech audio from text and save it to an audio file.

```
openrouter audio speak [text] [flags]
```

### Examples

```
  openrouter audio speak "Ship it, but make it calm." --output voice.mp3
  openrouter audio speak --model openai/gpt-4o-mini-tts-2025-12-15 --voice alloy --format mp3 "Hello from OpenRouter"
```

### Options

```
  -f, --file string       Read text from a file
      --force             Overwrite output files if they already exist
      --format string     Audio format: mp3 or pcm (default "mp3")
  -h, --help              help for speak
  -m, --model string      Speech model ID, or auto to choose a current speech model (default "auto")
      --output string     Output file path or directory (default: openrouter-audio-<timestamp>.<format>)
      --provider string   Provider routing JSON object
      --speed float       Playback speed multiplier for models that support it
      --stdin             Read text from stdin
  -t, --text string       Text to synthesize (can also be provided as positional args)
      --voice string      Voice identifier (default "alloy")
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

* [openrouter audio](openrouter_audio.md)	 - Generate speech and transcribe audio
