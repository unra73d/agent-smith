## Why

Users need to quickly locate text in a long chat without relying on the browser's page search, which does not provide chat-scoped navigation or match highlighting. A local find experience keeps searching responsive and avoids unnecessary backend requests.

## What Changes

- Add a VS Code-style find panel in the top-right of the active chat.
- Open the panel with Ctrl+F or Cmd+F and close it with its close control or Escape.
- Search only visible user messages and final model answers, literally and case-sensitively, without regular expressions, using a 500 ms typing debounce. Exclude tool calls, tool responses, and thinking/reasoning nodes, including when those nodes are expanded.
- Automatically cancel and close the active search when the user submits a new chat message, removing all highlights and pending search work.
- Highlight every match with a muted highlight, emphasize the active match with a distinct highlight, show the match count, and navigate matches with previous/next controls.
- Keep all searching and navigation in the UI; do not add backend calls or API changes.

## Capabilities

### New Capabilities
- `chat-text-search`: Provides local, navigable text search within the active chat session.

### Modified Capabilities
- None.

## Impact

- Affects the frontend `ChatView` component and its shadow-DOM styles.
- No backend, persistence, SSE, or external dependency changes.
- Search results must remain coherent when the active session is rebuilt or streamed content changes.
