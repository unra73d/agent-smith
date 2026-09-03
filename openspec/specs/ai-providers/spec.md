# AI Providers & Models Specification

## Purpose

Agent Smith connects to OpenAI-compatible LLM providers and exposes the models they serve to
the chat UI and agent. This capability covers provider and model lifecycle management,
connectivity testing, per-provider rate limiting, and both non-streaming and streaming chat
completions over the OpenAI wire format.

## Requirements

### REQ-AIP-001: Provider model
The system MUST represent an AI provider with a unique ID, display name, API base URL, API
key, an API type, an integer requests-per-minute rate limit, and a list of discovered models.

### REQ-AIP-002: Provider types
The system MUST model provider API types including OpenAI, LM Studio, Google, Mistral,
Ollama, Anthropic, and a generic OpenAI-compatible type, and MUST skip API-key
authentication headers for local providers (Ollama, LM Studio).

### REQ-AIP-003: Provider loading
The system MUST load all persisted providers from the `providers` table on startup,
reconstructing each provider and discovering its models asynchronously.

### REQ-AIP-004: Provider persistence
The system MUST persist providers using an upsert into a `providers` table keyed by ID,
storing id, name, API URL, API key, provider type, and rate limit.

### REQ-AIP-005: Provider deletion
The system MUST delete a provider by ID from storage and the in-memory list and MUST report
an error when the provider does not exist.

### REQ-AIP-006: Provider connectivity test
The system MUST test a provider's connectivity by attempting to load its model list, treating
the provider as reachable only when the model list request succeeds.

### REQ-AIP-007: Model discovery
The system MUST discover models by issuing a `GET <apiUrl>/models` request and mapping each
returned model ID to a `Model` entry referencing the owning provider, with a bounded request
timeout.

### REQ-AIP-008: Model list aggregation
The system MUST aggregate and expose the models of all providers as a single flattened list
of models.

### REQ-AIP-009: Provider create/update propagation
On provider create or update, the system MUST persist the change, reload the provider's model
list asynchronously when the URL or key changed, and broadcast a `provider_list_update` event.

### REQ-AIP-010: Rate limiting
The system MUST enforce a rolling one-minute rate limit on chat completion requests per
provider, blocking the caller until capacity is available; when a provider's rate limit is
zero, the system MUST apply no throttling.

### REQ-AIP-011: Non-streaming chat completion
The system MUST send a `POST <apiUrl>/chat/completions` request with the model ID and prepared
messages (a system message followed by conversation messages), and MUST return the content of
the first completion choice.

### REQ-AIP-012: Streaming chat completion
The system MUST support streaming chat completions over SSE, parsing OpenAI chunk deltas,
forwarding text content deltas to a write channel, accumulating tool-call deltas by index,
and retaining provider-reported completion-token usage from terminal stream metadata when
supplied, until the stream terminates with `[DONE]` or the context is cancelled.

#### Scenario: Stream text and usage
- **WHEN** an OpenAI-compatible provider streams text chunks followed by terminal completion-token usage metadata
- **THEN** the system MUST forward the text deltas and make the reported completion-token count available to the caller after successful stream completion

#### Scenario: Stream omits usage
- **WHEN** an OpenAI-compatible provider completes a stream without completion-token usage metadata
- **THEN** the system MUST complete normally and indicate that provider completion-token usage is unavailable

#### Scenario: Stream cancellation
- **WHEN** the streaming request context is cancelled
- **THEN** the system MUST stop processing the stream as a clean shutdown without reporting successful completion metadata

### REQ-AIP-013: Tool-call accumulation
When a provider streams tool calls, the system MUST accumulate fragmented function names and
arguments across chunks per call index, parse the final arguments JSON into a parameter map,
assign a call ID (generating one when absent), and emit the assembled tool call requests on a
tool channel.

### REQ-AIP-014: Message preparation
The system MUST prepare request messages by prepending the system prompt, setting empty
message content to the placeholder `<no response>`, stripping thinking content and trimming
text, and by attaching tool-call metadata for assistant and tool messages.

### REQ-AIP-015: Tool definitions
When tools are supplied, the system MUST serialize them into OpenAI function definitions with
name, description, typed properties, and required parameter lists, and attach them to the
completion request body.

### REQ-AIP-016: Graceful stream cancellation
The system MUST treat context cancellation during a streamed completion as a clean shutdown
rather than an error, allowing the caller to abort generation via request cancellation.
