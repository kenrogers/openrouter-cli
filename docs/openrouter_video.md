## openrouter video

Generate video from text and images

### Synopsis

Generate a video through OpenRouter's async video API.

By default the command submits the job, waits for completion, downloads the
first video, and returns the saved path. Use --no-wait to only submit and return
the job ID.

```
openrouter video [prompt] [flags]
```

### Examples

```
  openrouter video "a calm product shot of a glass keyboard on a walnut desk"
  openrouter video --model google/veo-3.1-lite --duration 4 --resolution 720p --aspect-ratio 16:9 "clouds over Denver"
  openrouter video --first-frame start.png --last-frame end.png --output clip.mp4 "animate this transition"
  openrouter video models veo
```

### Options

```
      --aspect-ratio string           Aspect ratio such as 16:9, 9:16, or 1:1
      --callback-url string           HTTPS callback URL for completion notification
      --duration int                  Duration in seconds
  -f, --file string                   Read prompt text from a file
      --first-frame string            First-frame image path, URL, or data URL
      --force                         Overwrite output files if they already exist
      --generate-audio                Generate audio when the selected model supports it
  -h, --help                          help for video
      --last-frame string             Last-frame image path, URL, or data URL
  -m, --model string                  Video model ID, or auto to choose a current video model (default "auto")
      --no-wait                       Submit the job and return without polling
      --output string                 Output file path or directory (default: openrouter-video-<timestamp>.mp4)
      --poll-interval duration        Polling interval while waiting for completion (default 5s)
  -p, --prompt string                 Video prompt (can also be provided as positional args)
      --provider string               Provider routing JSON object
      --reference-image stringArray   Reference image path, URL, or data URL; repeat for multiple references
      --resolution string             Resolution such as 480p, 720p, 1080p, 1K, 2K, or 4K
      --seed int                      Deterministic seed for models that support it
      --size string                   Exact size such as 1280x720
      --stdin                         Read prompt text from stdin
      --wait-timeout duration         Maximum time to wait for completion (default 15m0s)
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
* [openrouter video models](openrouter_video_models.md)	 - List video generation models
