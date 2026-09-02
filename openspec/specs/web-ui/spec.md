# Web UI Specification

## Purpose

Agent Smith ships a self-contained web UI served by the backend and displayed either in a
native webview or a browser. This capability describes the UI's layout and tabs, the chat
experience, reactive state management, the search features, and how the UI consumes the
backend API and real-time events.

## Requirements

### REQ-UI-001: Application layout
The UI MUST provide a top bar with tab buttons and model/role selectors, a collapsible side
panel containing the active tab's list, and a main chat view.

### REQ-UI-002: Navigation tabs
The UI MUST provide tabs for Chat Sessions, MCP servers/tools, Models (providers), and Roles,
each opening a corresponding panel; clicking the active tab MUST toggle the panel closed.

### REQ-UI-003: Reactive state store
The UI MUST maintain a client-side storage object for models, sessions, roles, MCP servers,
providers, and the current session, and MUST broadcast custom events whenever a monitored
collection changes so components re-render reactively.

### REQ-UI-004: Startup bootstrapping
On load, the UI MUST connect to the SSE endpoint and reload providers, sessions, MCP servers,
and roles, auto-creating a new session when no sessions exist.

### REQ-UI-005: Chat input
The UI MUST provide a chat textarea that auto-grows up to a maximum height, submits the
message on `Enter`, inserts a newline on `Shift+Enter`, and toggles the cancel/stop button
while generation is in progress.

### REQ-UI-006: Tools toggle
The UI MUST provide a "Tools" checkbox that selects between tool chat (agents + tools) and
direct chat (no tools) when sending a message.

### REQ-UI-007: Message rendering
The UI MUST render messages by origin, rendering assistant messages with collapsible thinking
and tool sections and markdown content, tool messages as collapsible tool responses, and user
messages as plain text.

### REQ-UI-008: Code block rendering and copy
The UI MUST render fenced code blocks with syntax highlighting, wrap each block with header
and footer copy buttons, and copy the block's plain text to the clipboard on click.

### REQ-UI-009: External link handling
The UI MUST intercept clicks on rendered links and open them through the backend's
`/desktop/url/open` endpoint instead of navigating the UI away.

### REQ-UI-010: Per-message actions
The UI MUST provide copy, reload (regenerate from the preceding user message), and delete
actions on user and assistant messages, with delete confirming via a dialog before calling the
backend.

### REQ-UI-011: Generation cancellation
The UI MUST track in-flight generation requests per session using `AbortController`s and MUST
abort the request when the user clicks Stop or a new generation is started for the same
session.

### REQ-UI-012: Chat text search
The UI MUST support finding text in the current chat via `Ctrl/Cmd+F`, highlighting all
matches (across user and assistant messages, excluding thinking/tool blocks), showing a match
count, and navigating between matches with `Enter`/`Shift+Enter` or the previous/next buttons,
debouncing input.

### REQ-UI-013: Session search
The UI MUST support filtering the session list by text, matching a session when the search
term appears in its title or in any of its message texts.

### REQ-UI-014: Real-time message updates
The UI MUST react to `new_message`, `last_message_update`, and `session_update` SSE events by
appending messages, updating the in-progress assistant message, and replacing updated sessions.

### REQ-UI-015: Session list management
The UI MUST support creating, deleting, and selecting sessions from the list, auto-selecting
a remaining session after deletion, and moving the current session to the top when it is
touched (a new message is sent).

### REQ-UI-016: Model and role selection
The UI MUST populate the model selector from all providers' models and the role selector from
the roles list, reflecting changes delivered via the real-time provider and role list update
events.