## Why

Completed AI responses currently expose neither their generation speed nor duration. KAN-5 makes that performance information visible directly beneath the final visible assistant answer without interrupting streamed content.

## What Changes

- Record output-token counts and elapsed generation durations for successful final assistant answers.
- Use provider-reported completion-token usage when available; otherwise produce a local output-token estimate.
- Preserve message statistics through the existing session persistence and live SSE message-update path.
- Render compact muted statistics beneath each completed final assistant answer in the form `30t/s 5m 33s`, reserving footer space so its hover-only visibility does not shift surrounding messages.
- Omit statistics for assistant messages that fail or are cancelled, and for intermediate tool-call turns.

## Capabilities

### New Capabilities
- `ai-message-statistics`: Captures and displays performance statistics for successful final assistant answers.

### Modified Capabilities
- `ai-providers`: Streaming completions expose terminal completion-token usage when the provider supplies it.
- `sessions`: Chat messages persist optional response statistics and deliver final statistics through the existing last-message update event.
- `agent-orchestration`: Direct and tool chat associate successful final answers, but not intermediate tool turns, with response statistics.
- `web-ui`: Assistant-message rendering displays completed response statistics.

## Impact

The `ai` package gains streaming usage capture and message statistics data; `agent` attaches statistics to final successful answers before session persistence; the existing SSE payload automatically carries the expanded message shape; and the chat web component renders the footer. No new route, SSE event type, database table, or dependency is introduced. Legacy message JSON remains compatible because statistics are optional.
