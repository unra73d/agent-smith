## Why

Users with many chat sessions cannot find a specific conversation. The chat
history list in the side panel offers no way to filter conversations, so
locating an old chat requires scrolling through every session and guessing from
summaries.

## What Changes

- Add a single-line search field at the top of the chat sessions list (the
  "Chat sessions" tab in the side panel).
- As the user types, the chat list filters in real time.
- A chat matches when the query appears in the chat title (session summary) or
  in the text of any message inside the chat. Matching is case-insensitive.
- A clear ("x") button inside the field lets the user reset the search and
  restore the full list instantly.
- The search field matches the app's dark theme and the styling of the other
  text inputs (same dark background, border, and text colors); the clear button
  sits on the same line as the field (inside the input, vertically centered).
  This addresses the review feedback that the field rendered white (default
  browser style) with the clear button wrapping to the next line.

## Capabilities

### New Capabilities
- `chat-history-search`: The chat history (sessions list) can be filtered by a
  search query typed into a dedicated field at the top of the list. The filter
  matches against the chat title or the text of messages inside the chat, runs
  in real time as the user types, and can be cleared with an inline "x" button.

### Modified Capabilities
<!-- No existing requirement changes; this is a new UI behavior. -->

## Impact

- Frontend only (no backend/API changes): the session list already carries
  `summary` and `messages` client-side, so filtering is purely a UI concern.
- `src/ui/index.html`: add the search field markup to the Chat sessions tab.
- `src/ui/components/sessions/sessions.js`: add filter state and real-time
  filtering of the session list.
- `src/ui/styles.css`: style the search field, inline clear button, and
  active-filter states in the main document stylesheet (the field markup lives
  in the light DOM, so its styles must be loaded by the main document, not the
  `session-list` shadow root), and adjust the session list max-height so the
  list remains scrollable with the search bar pinned above it.
