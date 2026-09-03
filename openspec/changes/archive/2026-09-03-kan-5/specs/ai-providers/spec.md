## MODIFIED Requirements

### Requirement: Streaming chat completion
The system MUST support streaming chat completions over SSE, parsing OpenAI chunk deltas, forwarding text content deltas to a write channel, accumulating tool-call deltas by index, and retaining provider-reported completion-token usage from terminal stream metadata when supplied, until the stream terminates with `[DONE]` or the context is cancelled.

#### Scenario: Stream text and usage
- **WHEN** an OpenAI-compatible provider streams text chunks followed by terminal completion-token usage metadata
- **THEN** the system MUST forward the text deltas and make the reported completion-token count available to the caller after successful stream completion

#### Scenario: Stream omits usage
- **WHEN** an OpenAI-compatible provider completes a stream without completion-token usage metadata
- **THEN** the system MUST complete normally and indicate that provider completion-token usage is unavailable

#### Scenario: Stream cancellation
- **WHEN** the streaming request context is cancelled
- **THEN** the system MUST stop processing the stream as a clean shutdown without reporting successful completion metadata
