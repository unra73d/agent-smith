## MODIFIED Requirements

### Requirement: Direct chat session semantics
Direct chat MUST append a user message and an empty assistant message to the session, stream text into the assistant message, persist the session, and trigger title generation on completion. After a successful completion, it MUST record response statistics on that assistant answer; on failure or context cancellation, it MUST not record response statistics. It MUST signal stream completion/failure to the caller via a completion channel.

#### Scenario: Successful direct answer
- **WHEN** a direct chat response completes successfully
- **THEN** the system MUST record and persist response statistics on its assistant answer before notifying connected clients of the final message update

#### Scenario: Failed or cancelled direct answer
- **WHEN** a direct chat response fails or its context is cancelled
- **THEN** the system MUST not record response statistics for its assistant message

### Requirement: Answer termination
When the model decides to answer directly, the system MUST record response statistics only for that successful final assistant answer, trigger title generation, signal successful stream completion, and stop the tool loop. It MUST not record or display response statistics for assistant turns that result in tool calls.

#### Scenario: Tool-assisted final answer
- **WHEN** a tool chat contains one or more intermediate assistant tool-call turns followed by a successful final answer
- **THEN** the system MUST record response statistics only on the final answer
