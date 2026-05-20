## openrouter beta responses send

Create a response

### Synopsis

Creates a streaming or non-streaming response using OpenResponses API format

```
openrouter beta responses send [flags]
```

### Examples

```
  openrouter responses send
```

### Options

```
  -b, --background string                                         boolean flag
      --body string                                               Request body as JSON (alternative to individual flags). Can also be provided via stdin.
  -c, --cache-control string                                      Enable automatic prompt caching. When set at the top level, the system automatically applies cache breakpoints to the last cacheable block in the request. Currently supported for Anthropic Claude models.
  -f, --frequency-penalty string                                  number value
  -h, --help                                                      help for send
      --image-config string                                       Provider-specific image configuration options. Keys and values vary by model/provider. See https://openrouter.ai/docs/guides/overview/multimodal/image-generation for more details.
      --include string                                            list of values
      --input string                                              JSON value (one of: string | array of { content: object[], encrypted_content: string, id: string, status: value, ... } | { content: value, phase: value, role: value, type: string } | { content: value[], id: string, role: value, type: string } | { arguments: string, call_id: string, id: string, name: string, ... } | { call_id: string, id: string, output: value, status: string, ... } | { call_id: string, id: string, operation: value, status: string, ... } | { call_id: string, id: string, output: string, status: string, ... } | { content: value, id: string, phase: value, role: string, ... } | { call_id: string, id: string, input: string, name: string, ... } | { action: value, id: string, status: string, type: string } | { id: string, queries: string[], status: string, type: string } | { id: string, result: string, status: string, type: string } | { code: string, container_id: string, id: string, outputs: value[], ... } | { action: value, call_id: string, id: string, pending_safety_checks: object[], ... } | { datetime: string, id: string, status: string, timezone: string, ... } | { action: object, id: string, status: string, type: string } | { code: string, exitCode: integer, id: string, language: string, ... } | { id: string, imageB64: string, imageUrl: string, result: string, ... } | { action: string, id: string, screenshotB64: string, status: string, ... } | { command: string, exitCode: integer, id: string, status: string, ... } | { command: string, filePath: string, id: string, status: string, ... } | { content: string, error: string, httpStatus: integer, id: string, ... } | { id: string, query: string, status: string, type: string } | { action: string, id: string, key: string, status: string, ... } | { id: string, serverLabel: string, status: string, toolName: string, ... } | { arguments: string, id: string, query: string, status: string, ... } | { action: object, call_id: string, id: string, status: string, ... } | { id: string, output: string, status: string, type: string } | { action: object, call_id: string, environment: value, id: string, ... } | { call_id: string, id: string, max_output_length: integer, output: object[], ... } | { error: string, id: string, server_label: string, tools: object[], ... } | { arguments: string, id: string, name: string, server_label: string, ... } | { approval_request_id: string, approve: boolean, id: string, reason: string, ... } | { arguments: string, error: string, id: string, name: string, ... } | { call_id: string, id: string, output: value, type: string } | { encrypted_content: string, id: string, type: string } | { id: string, type: string })
      --instructions string                                       string value
      --max-output-tokens string                                  integer value
      --max-tool-calls string                                     integer value
      --metadata string                                           Metadata key-value pairs for the request. Keys must be ≤64 characters and cannot contain brackets. Values must be ≤512 characters. Maximum 16 pairs allowed.
      --modalities stringArray                                    Output modalities for the response. Supported values are "text" and "image".
      --model string                                              string value
      --models stringArray                                        list of values
      --parallel-tool-calls string                                boolean flag
      --plugins string                                            Plugins you want to enable for this request, including their settings.
      --presence-penalty string                                   number value
      --previous-response-id string                               string value
      --prompt string                                             JSON object
      --prompt-cache-key string                                   string value
      --provider string                                           When multiple model providers are available, optionally indicate your routing preference.
  -r, --reasoning string                                          Configuration for reasoning mode in the response
      --safety-identifier string                                  string value
      --service-tier string                                       options: auto, default, flex, priority, scale (default "auto")
      --session-id string                                         A unique identifier for grouping related requests (e.g., a conversation or agent workflow) for observability. If provided in both the request body and the x-session-id header, the body value takes precedence. Maximum of 256 characters.
      --stop-server-tools-when max_tool_calls                     Stop conditions for the server-tool agent loop. Any condition firing halts the loop (OR logic). When set, this overrides max_tool_calls.
      --stream                                                    boolean flag
      --temperature string                                        number value
      --text string                                               Text output configuration including format and verbosity
      --tool-choice string                                        JSON value (one of: OpenAIResponsesToolChoice_1 | OpenAIResponsesToolChoice_2 | OpenAIResponsesToolChoice_3 | { name: string, type: string } | { type: value } | { mode: value, tools: object[], type: string } | { type: string })
      --tools string                                              list of values
      --top-k int                                                 integer value
      --top-logprobs string                                       integer value
      --top-p string                                              number value
      --trace string                                              Metadata for observability and tracing. Known keys (trace_id, trace_name, span_name, generation_name, parent_span_id) have special handling. Additional keys are passed through as custom metadata to configured broadcast destinations.
      --truncation string                                         options: auto, disabled
  -u, --user string                                               A unique identifier representing your end-user, which helps distinguish between different users of your app. This allows your app to identify specific users in case of abuse reports, preventing your entire app from being affected by the actions of individual users. Maximum of 256 characters.
  -x, --x-open-router-experimental-metadata openrouter_metadata   Opt-in to surface routing metadata on the response under openrouter_metadata. Defaults to `disabled`. (options: disabled, enabled)
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

* [openrouter beta responses](openrouter_beta_responses.md)	 - beta
