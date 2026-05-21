## openrouter image

Generate or edit images

### Synopsis

Generate or edit images through OpenRouter image-output models.

This is the friendly path for agents and developers. It wraps the underlying
chat-completions image workflow, saves returned base64 image data to files, and
returns machine-readable output without requiring hand-built JSON.

```
openrouter image [prompt] [flags]
```

### Examples

```
  openrouter image "a tiny red robot, product photo style"
  openrouter image --model google/gemini-3.1-flash-image-preview --aspect-ratio 16:9 --output hero.png "a cinematic mountain sunrise"
  openrouter image --input-image avatar.png --output avatar-watercolor.png "turn this into a watercolor portrait"
  openrouter image models nano banana
```

### Options

```
      --aspect-ratio string       Requested aspect ratio, such as 1:1, 16:9, 9:16, or 4:1
      --force                     Overwrite output files if they already exist
  -h, --help                      help for image
      --image-config string       Advanced image_config JSON object; explicit image flags override matching keys
      --image-size string         Requested image size, such as 0.5K, 1K, 2K, or 4K
      --input-image stringArray   Input image path, URL, or data URL for image editing; repeat for multiple images
  -m, --model string              Image-capable model ID, or auto to choose a current image-output model (default "auto")
      --output string             Output file path or directory (default: openrouter-image-<timestamp>.<ext>)
  -p, --prompt string             Image generation or edit prompt (can also be provided as positional args)
      --strength float            Image edit strength for models that support it (0.0 to 1.0) (default -1)
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
* [openrouter image models](openrouter_image_models.md)	 - List image generation models
