## openrouter video-generation generate

Submit a video generation request

### Synopsis

Submits a video generation request and returns a polling URL to check status

```
openrouter video-generation generate [flags]
```

### Examples

```
  openrouter video-generation generate --model google/veo-3.1 --prompt A serene mountain landscape at sunset
```

### Options

```
  -a, --aspect-ratio string       Aspect ratio of the generated video (options: 16:9, 9:16, 1:1, 4:3, 3:4, 3:2, 2:3, 21:9, 9:21)
      --body string               Request body as JSON (alternative to individual flags). Can also be provided via stdin.
  -c, --callback-url string       URL to receive a webhook notification when the video generation job completes. Overrides the workspace-level default callback URL if set. Must be HTTPS.
      --duration int              Duration of the generated video in seconds
  -f, --frame-images string       Images to use as the first and/or last frame of the generated video. Each image must specify a frame_type of first_frame or last_frame.
  -g, --generate-audio            Whether to generate audio alongside the video. Defaults to the endpoint's generate_audio capability flag, false if not set.
  -h, --help                      help for generate
  -i, --input-references string   Reference images to guide video generation
  -m, --model string              [required]
      --prompt string             [required]
      --provider string           Provider-specific passthrough configuration
  -r, --resolution string         Resolution of the generated video (options: 480p, 720p, 1080p, 1K, 2K, 4K)
      --seed int                  If specified, the generation will sample deterministically, such that repeated requests with the same seed and parameters should return the same result. Determinism is not guaranteed for all providers.
      --size string               Exact pixel dimensions of the generated video in "WIDTHxHEIGHT" format (e.g. "1280x720"). Interchangeable with resolution + aspect_ratio.
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

* [openrouter video-generation](openrouter_video-generation.md)	 - Video Generation endpoints
