# Chat Sessions Specification

## Purpose

A chat session is the primary unit of conversational state. This capability covers the
session and message lifecycle: creating, listing, deleting and truncating sessions; adding,
updating and deleting messages; persisting sessions to SQLite; and asynchronously generating
human-readable session titles from conversation content.

## Requirements

### REQ-SES-001: Session model
The system MUST represent a session with a unique identifier (`ID`), a creation/update
timestamp (`Date`), an ordered list of messages (`Messages`), and a human-readable summary
(`Summary`), and MUST mark certain sessions as temporary (non-persisted).

### REQ-SES-002: Session creation
The system MUST create a new session with a fresh UUID, the current timestamp, an empty
message list, and the default summary "New chat", persist it, and make it the newest entry in
the in-memory session list.

### REQ-SES-003: Session listing
The system MUST load and return all persisted sessions ordered by date descending, restoring
each session's summary and its JSON-encoded message list from storage.

### REQ-SES-004: Session deletion
The system MUST delete a session by ID from both storage and the in-memory list, and MUST
report an error when the requested session does not exist.

### REQ-SES-005: Message model
The system MUST represent a chat message with a unique identifier, an origin role (user,
assistant, tool, or system), free-text content, and an optional list of tool call requests.

### REQ-SES-006: Add message
The system MUST append a new message to a session, refresh the session timestamp, broadcast a
`new_message` and a `session_update` event, and persist the session unless it is temporary.

### REQ-SES-007: Streaming last-message update
The system MUST support incrementally appending text to the last message of a session and
broadcast a `last_message_update` event carrying the session ID and the updated message.

### REQ-SES-008: Delete message
The system MUST delete a message from a session. When the deleted message is an assistant
message, the system MUST also remove the backwards chain of tool requests and tool result
messages up to and including the preceding user message. The system MUST broadcast a
`session_update` event and persist after deletion.

### REQ-SES-009: Truncate session
The system MUST truncate a session so that it retains only messages before a given message
ID, persist the change, and broadcast a `session_update` event.

### REQ-SES-010: Session persistence
The system MUST persist session state using an upsert (`INSERT ... ON CONFLICT DO UPDATE`)
into a `sessions` table keyed by session ID, storing the date, summary, and the JSON-encoded
message list, and MUST skip persistence for temporary sessions.

### REQ-SES-011: Temporary session
The system MUST provide a way to create non-persisted temporary sessions used for one-shot
agent operations that must not appear in history.

### REQ-SES-012: Title generation eligibility
The system MUST attempt to (re)generate a session title only when the session is not
temporary, has between one and three user messages, has a non-nil model, and has at least one
user or assistant message with non-empty trimmed text.

### REQ-SES-013: Title generation input filtering
When generating a title, the system MUST include only user and assistant messages with
non-empty trimmed text, flattening them into a single user turn that ends on a user prompt
asking for the title, and MUST do so asynchronously so it never delays the chat response.

### REQ-SES-014: Title sanitization
The system MUST sanitize a generated title by stripping leading thinking/reasoning content,
trimming whitespace and surrounding quotes, and, when the remaining text is too long to be a
2-4 word title, falling back to the last non-empty paragraph.

### REQ-SES-015: Title persistence and broadcast
On successful title generation, the system MUST update the session summary, persist it, and
broadcast a `session_update` event so connected clients update the session list and header.