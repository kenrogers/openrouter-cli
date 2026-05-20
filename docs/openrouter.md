## openrouter

OpenRouter API: OpenAI-compatible API with additional OpenRouter features

### Synopsis

OpenRouter API: OpenAI-compatible API with additional OpenRouter features

```
openrouter [flags]
```

### Options

```
      --agent-mode                      Enable structured errors and default TOON output for AI coding agents. Automatically enabled when a known agent environment is detected (CLAUDE_CODE, CURSOR_AGENT, etc.). Use --agent-mode=false to disable.
      --api-key string                  API key as bearer token in Authorization header
      --app-categories string           Comma-separated list of app categories (e (env: OPENROUTER_APP_CATEGORIES)
      --app-title string                The app display name allows you to customize how your app appears in OpenRouter's dashboard (env: OPENROUTER_APP_TITLE)
      --color string                    Control colored output: auto (color when output is a TTY), always, or never. Respects NO_COLOR and FORCE_COLOR env vars. (default "auto")
  -d, --debug                           Log request and response diagnostics to stderr
      --dry-run                         Preview the request that would be sent without executing it (output to stderr)
  -H, --header stringArray              Set a custom HTTP request header (format: "Key: Value"). Can be specified multiple times.
  -h, --help                            help for openrouter
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

* [openrouter O-auth](openrouter_O-auth.md)	 - OAuth authentication endpoints
* [openrouter analytics](openrouter_analytics.md)	 - Analytics and usage endpoints
* [openrouter auth](openrouter_auth.md)	 - Manage authentication credentials
* [openrouter beta](openrouter_beta.md)	 - Operations for beta
* [openrouter byok](openrouter_byok.md)	 - BYOK endpoints
* [openrouter chat](openrouter_chat.md)	 - Operations for chat
* [openrouter configure](openrouter_configure.md)	 - Configure authentication, global parameters, and preferences
* [openrouter credits](openrouter_credits.md)	 - Credit management endpoints
* [openrouter doctor](openrouter_doctor.md)	 - Check OpenRouter CLI authentication and configuration
* [openrouter embeddings](openrouter_embeddings.md)	 - Text embedding endpoints
* [openrouter endpoints](openrouter_endpoints.md)	 - Endpoint information
* [openrouter exec](openrouter_exec.md)	 - Run a command with OPENROUTER_API_KEY injected from secure storage
* [openrouter explore](openrouter_explore.md)	 - Interactively browse and run commands
* [openrouter generations](openrouter_generations.md)	 - Generation history endpoints
* [openrouter guardrails](openrouter_guardrails.md)	 - Guardrails endpoints
* [openrouter keys](openrouter_keys.md)	 - API key management endpoints
* [openrouter login](openrouter_login.md)	 - Authenticate with OpenRouter
* [openrouter logout](openrouter_logout.md)	 - Clear saved OpenRouter credentials
* [openrouter models](openrouter_models.md)	 - Model information endpoints
* [openrouter observability](openrouter_observability.md)	 - Observability endpoints
* [openrouter organization](openrouter_organization.md)	 - Organization endpoints
* [openrouter presets](openrouter_presets.md)	 - Presets endpoints
* [openrouter providers](openrouter_providers.md)	 - Provider information endpoints
* [openrouter rerank](openrouter_rerank.md)	 - Rerank endpoints
* [openrouter stt](openrouter_stt.md)	 - Speech-to-text endpoints
* [openrouter tts](openrouter_tts.md)	 - Text-to-speech endpoints
* [openrouter version](openrouter_version.md)	 - Print the CLI version
* [openrouter video-generation](openrouter_video-generation.md)	 - Video Generation endpoints
* [openrouter whoami](openrouter_whoami.md)	 - Display current authentication and global parameter configuration
* [openrouter workspaces](openrouter_workspaces.md)	 - Workspaces endpoints
