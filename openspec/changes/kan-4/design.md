## Context

`ChatView` owns the chat shadow DOM and rebuilds message markup when sessions change or assistant output streams. Assistant messages are rendered from Markdown and syntax-highlighted, so search decoration must be applied after rendering and be safe to repeat.

## Goals / Non-Goals

**Goals:**
- Keep find UI and state inside `ChatView`.
- Search rendered message text without changing the stored session or API contract.
- Preserve existing Markdown links, code markup, collapsible blocks, and message actions while adding highlights to eligible text only.

**Non-Goals:**
- Backend/full-text search, persistence of queries, regex/fuzzy matching, or searching across sessions.

## Decisions

- Use a small panel in the chat shadow root with an input, count, previous/next controls, and close control. This keeps positioning and DOM access local to the chat component.
- Handle Ctrl/Cmd+F at the document level in `ChatView`, ignore events originating in editable fields, prevent the default browser action, and focus/select the panel input.
- Debounce matching by 500 ms. This is long enough to avoid work per keystroke while retaining the expected find-as-you-type feel.
- Traverse text nodes in rendered user message content and final `.message-content` model-answer content rather than replacing whole message `innerHTML`. Do not traverse `.thinking-block` or `.tool-block`, regardless of their open/closed state. Wrap matching ranges in `mark` elements so existing links, code spans, and controls remain intact. Exclude the search panel itself from traversal.
- Track the ordered match elements in component state. Navigation updates only the active class and scrolls the selected mark into view; rebuilding a message clears/reapplies marks and resets the active match when necessary.
- Reapply the active query after session changes and last-message streaming updates. Clear marks before each pass to avoid nested or stale highlights.
- Cancel search synchronously at the start of `sendMessageStreaming`, clearing the debounce timer, removing marks, resetting match state, and hiding the panel before the new message is sent.

## Risks / Trade-offs

- [Risk] Repeated DOM decoration can conflict with streaming rerenders → centralize clear-and-apply logic and invoke it after every render path.
- [Risk] Text-node wrapping can split nodes around Markdown syntax → use a TreeWalker over text nodes and skip existing marks, controls, and the find panel.
- [Risk] A query may become stale while the debounce timer is pending → cancel the prior timer and validate the latest query before applying results.

## Migration Plan

No data or API migration is required. The change is fully backward-compatible and can be rolled back by removing the `ChatView` find UI, state, and styles.
