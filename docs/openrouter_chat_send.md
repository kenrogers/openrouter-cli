## openrouter chat send

Create a chat completion

### Synopsis

Sends a request for a model response for the given chat conversation. Supports both streaming and non-streaming modes.

```
openrouter chat send [flags]
```

### Examples

```
  openrouter chat send --messages '[{"content":"You are a helpful assistant.","role":"system"},{"content":"What is the capital of France?","role":"user"}]'
```

### Options

```
      --body string                                               Request body as JSON (alternative to individual flags). Can also be provided via stdin.
  -c, --cache-control string                                      Enable automatic prompt caching. When set at the top level, the system automatically applies cache breakpoints to the last cacheable block in the request. Currently supported for Anthropic Claude models.
      --debug-param string                                        Debug options for inspecting request transformations (streaming only)
  -f, --frequency-penalty string                                  Frequency penalty (-2.0 to 2.0)
  -h, --help                                                      help for send
  -i, --image-config string                                       Provider-specific image configuration options. Keys and values vary by model/provider. See https://openrouter.ai/docs/guides/overview/multimodal/image-generation for more details.
      --logit-bias string                                         Token logit bias adjustments
      --logprobs string                                           Return log probabilities
      --max-completion-tokens string                              Maximum tokens in completion
      --max-tokens string                                         Maximum tokens (deprecated, use max_completion_tokens). Note: some providers enforce a minimum of 16.
      --messages string                                           List of messages for the conversation [required]
      --metadata string                                           Key-value pairs for additional object information (max 16 pairs, 64 char keys, 512 char values)
      --modalities stringArray                                    Output modalities for the response. Supported values are "text", "image", and "audio".
      --model string                                              Model to use for completion
      --models stringArray                                        Models to use for completion
      --parallel-tool-calls string                                Whether to enable parallel function calling during tool use. When true, the model may generate multiple tool calls in a single response.
      --plugins string                                            Plugins you want to enable for this request, including their settings.
      --presence-penalty string                                   Presence penalty (-2.0 to 2.0)
      --provider string                                           When multiple model providers are available, optionally indicate your routing preference.
      --reasoning string                                          Configuration options for reasoning models
      --response-format string                                    JSON value (variants: grammar: { grammar: string, type: string }, json_object: { type: string }, json_schema: { json_schema: object, type: string }, python: { type: string }, text: { type: string })
      --response-format.grammar string                            ChatFormatGrammarConfig variant as JSON
      --response-format.grammar.grammar string                    Custom grammar for text generation [required]
      --response-format.json-object string                        FormatJsonObjectConfig variant as JSON
      --response-format.json-schema string                        ChatFormatJsonSchemaConfig variant as JSON
      --response-format.python string                             ChatFormatPythonConfig variant as JSON
      --response-format.text string                               ChatFormatTextConfig variant as JSON
      --seed string                                               Random seed for deterministic outputs
      --service-tier string                                       The service tier to use for processing this request. (options: auto, default, flex, priority, scale)
      --session-id string                                         A unique identifier for grouping related requests (e.g., a conversation or agent workflow) for observability. If provided in both the request body and the x-session-id header, the body value takes precedence. Maximum of 256 characters.
      --stop string                                               JSON value (one of: string | array of string | any)
      --stop-server-tools-when max_tool_calls                     Stop conditions for the server-tool agent loop. Any condition firing halts the loop (OR logic). When set, this overrides max_tool_calls.
      --stream                                                    Enable streaming response
      --stream-options string                                     Streaming configuration options
      --temperature string                                        Sampling temperature (0-2)
      --tool-choice string                                        JSON value (one of: ChatToolChoice_1 | ChatToolChoice_2 | ChatToolChoice_3 | { function: object, type: string } | { type: string })
      --tools string                                              Available tools for function calling
      --top-logprobs string                                       Number of top log probabilities to return (0-20)
      --top-p string                                              Nucleus sampling parameter (0-1)
      --trace string                                              Metadata for observability and tracing. Known keys (trace_id, trace_name, span_name, generation_name, parent_span_id) have special handling. Additional keys are passed through as custom metadata to configured broadcast destinations.
  -u, --user string                                               Unique user identifier
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

* [openrouter chat](openrouter_chat.md)	 - Operations for chat
