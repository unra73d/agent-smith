## Why

When a user starts a chat, the session's `Summary` (shown as the title in the left-hand chat list) is currently just the raw text of the first user message, truncated at whatever length it happens to be. This is often unreadable or misleading, especially for long first messages or when the conversation's real topic only becomes clear after a reply or two. KAN-3 asks for a real human-readable title, generated from the conversation content, that keeps updating for the first few exchanges since users often "test" a chat with a throwaway first message before getting to the real topic.

## What Changes

- Replace the naive `Summary = first message text` assignment in `Session.AddMessage` with model-generated title summarization.
- After each of the first 3 completed user/assistant exchanges (an exchange completes when the assistant produces its final answer for that turn, with no pending tool calls), regenerate a short 2-4 word title from the conversation so far and persist/broadcast it.
- After the 3rd exchange, stop regenerating — the title stabilizes.
- Title generation input is limited to user and assistant message text; tool-call/tool-result messages and empty/whitespace-only messages are excluded.
- Title generation runs asynchronously (a goroutine) after the assistant's reply is sent, so it never delays the chat response.
- Temporary/flash sessions are never titled or persisted (unchanged from current behavior, made explicit).
- No frontend changes: the existing `session_update` SSE event and `sessions.js` rendering/search already handle `session.summary` correctly.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `chat-sessions`: adds a requirement that session titles are auto-generated and progressively refined from conversation content during the first 3 exchanges, instead of being copied verbatim from the first message.

## Impact

- `src/agent/session.go`: `AddMessage` no longer sets `Summary` from raw first-message text; new helper for exchange counting and title assignment.
- `src/agent/tool_agent.go`: `ToolChatStreaming` triggers title (re)generation at the `AgentActionAnswer` completion point.
- `src/agent/chat_agent.go`: `DirectChatStreaming` triggers title (re)generation at reply completion.
- `src/ai/apiprovider.go`: reuses existing `ChatCompletion` for the title-generation model call.
- No database schema changes (`summary` column already exists).
- No frontend changes.
