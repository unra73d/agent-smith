# Agent Orchestration Specification

## Purpose

The agent orchestrates requests between the LLM and its tools. This capability covers the
three chat modes (direct chat, tool/agent chat, and dynamic agent chat), how the agent
infers whether the model intends to answer or call a tool, how tool results are fed back into
the model loop, and how recursion is bounded.

## Requirements

### REQ-AGO-001: Direct chat streaming
The system MUST support a direct chat mode that sends the user message to the model with a
role-derived system prompt and no tools, streaming the model's text deltas into the session's
last message until completion.

### REQ-AGO-002: Direct chat session semantics
Direct chat MUST append a user message and an empty assistant message to the session, stream
text into the assistant message, persist the session, and trigger title generation on
completion. After a successful completion, it MUST record response statistics on that assistant
answer; on failure or context cancellation, it MUST not record response statistics. It MUST
signal stream completion/failure to the caller via a completion channel.

#### Scenario: Successful direct answer
- **WHEN** a direct chat response completes successfully
- **THEN** the system MUST record and persist response statistics on its assistant answer before notifying connected clients of the final message update

#### Scenario: Failed or cancelled direct answer
- **WHEN** a direct chat response fails or its context is cancelled
- **THEN** the system MUST not record response statistics for its assistant message

### REQ-AGO-003: Model resolution
The system MUST resolve a model by ID across all providers and MUST abort the chat operation
with a failure signal when the requested model cannot be found.

### REQ-AGO-004: Role prompt composition
When a role is supplied, the system MUST compose the system prompt from the role's general
instruction, role/personality, and text style sections; when no role matches, the system MUST
use an empty system prompt.

### REQ-AGO-005: Tool chat streaming
The system MUST support a tool chat mode in which the model may repeatedly request tool calls,
each tool result is appended as a tool message, and the model is re-invoked until it decides
to answer directly.

### REQ-AGO-006: Tool availability
Tool chat MUST expose the active MCP tools plus the built-in tools to the model, and MUST
append the tool-usage prompt to the system prompt only when at least one tool is available.

### REQ-AGO-007: Action classification
After each model turn, the system MUST classify the model's intent into one of: answer,
tool-call, or error. A tool-call is recognized from structured tool calls emitted by the
provider or by parsing the last message text for a `<tool_call>` JSON block.

### REQ-AGO-008: Tool name matching
When a tool call is parsed from text, the system MUST match the requested tool name to an
available tool using an exact match first, then a substring match to tolerate models that
prefix or suffix tool names, and MUST normalize the call name to the matched tool's name.

### REQ-AGO-009: Tool execution
When a tool call is requested, the system MUST record the tool call on the last assistant
message and execute the tool through its owning MCP server, appending the sanitized result as
a tool message, or appending an error tool message when execution fails, then re-invoking the
model.

### REQ-AGO-010: Tool execution failure resilience
When a tool call fails, the system MUST NOT terminate the stream; instead it MUST append a
tool message describing the failure so the model can react to it.

### REQ-AGO-011: Unknown tool handling
When the model requests a tool that cannot be resolved to a server (except the built-in Lua
runner), the system MUST signal stream failure and abort.

### REQ-AGO-012: Answer termination
When the model decides to answer directly, the system MUST record response statistics only for
that successful final assistant answer, trigger title generation, signal successful stream
completion, and stop the tool loop. It MUST not record or display response statistics for
assistant turns that result in tool calls.

#### Scenario: Tool-assisted final answer
- **WHEN** a tool chat contains one or more intermediate assistant tool-call turns followed by a successful final answer
- **THEN** the system MUST record response statistics only on the final answer

### REQ-AGO-013: Thinking content stripping
Before parsing tool calls or persisting tool results, the system MUST strip thinking/reasoning
tags (`<think>`, `<thinking>`) from message text and trim the remaining content.

### REQ-AGO-014: Dynamic agent chat
The system MUST support a one-shot dynamic agent mode that runs a non-persisted temporary
session with a caller-supplied system prompt and no tools, returns the complete model response,
and does not save any messages or sessions.

### REQ-AGO-015: Context cancellation
The system MUST abort active generation and signal failure when the request context is
cancelled, enabling manual stop of generation.
