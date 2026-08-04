## Context

The chat history is rendered by the `session-list` custom element
(`src/ui/components/sessions/sessions.js`), a subclass of the shared `List`
component (`src/ui/components/list/list.js`). `List` re-renders whenever its
`items` property changes (via a `list:change` event). Each session object
already carries a `summary` (the displayed chat title) and a `messages` array
of `{ text, origin, toolRequests }` objects, so the data needed to filter
already lives client-side — no backend change is required. The side panel tab
markup lives in `src/ui/index.html`; shared styles live in `src/ui/styles.css`.

## Goals / Non-Goals

**Goals:**
- Add a single-line search field pinned at the top of the Chat sessions tab.
- Filter the rendered session list in real time, case-insensitively, against
  the chat title (session summary) or any message text.
- Provide an inline clear ("x") button that resets the search.
- Keep the change scoped to the sessions tab so the other tabs (MCP, Models,
  Roles) are untouched.

**Non-Goals:**
- Server-side search or a dedicated search API.
- Fuzzy/ranked search, highlighting of matched text, or debounce optimization
  (list sizes are small; direct filtering is sufficient).
- Applying the search field to the other tabs' lists.

## Decisions

- **Filter in the `SessionList` subclass, not the shared `List` base.**
  Overriding `updateList()` in `SessionList` to skip non-matching sessions keeps
  the shared `List` component (used by MCP/providers/roles tabs) unchanged and
  avoids the risk of regressing those tabs.

- **Keep `items` as the full session list and filter at render time.**
  The base component already re-renders on `list:change` when `Storage.sessions`
  changes (create/delete/touch/SSE updates). Filtering at render time means the
  search stays active across session updates without extra bookkeeping.

- **Search field lives in the light DOM in `index.html`.**
  The input is regular markup in the Chat sessions tab, referenced from the
  `SessionList` constructor via `document.getElementById`. The field's `input`
  event updates `this.filter` and re-renders the list. The clear button is an
  absolutely-positioned button inside a relatively-positioned wrapper, shown
  only while the field has text.

- **Match against the displayed title and raw message text.**
  The list shows `session.summary` (or "New chat" when empty), so matching uses
  the same fallback. Message matching iterates `session.messages` and checks
  `msg.text`. Case-insensitivity via `.toLowerCase()` on both sides.

- **Reuse existing CSS patterns.**
  Colors, borders, and the `img-button` icon styling follow the existing dark
  theme (`#313335` input backgrounds, `#3c4043` borders, `#8ab4f8` focus).

## Risks / Trade-offs

- [List max-height may overflow with the added search bar] → Adjust the
  `.session-list` max-height in `src/ui/styles.css` so the list remains its own
  scroll area with the search bar pinned above it.
- [`document.getElementById` from a custom element constructor couples the
  component to the light DOM] → The app already relies on `index.html` markup
  order; the search field is static markup, so the reference resolves at
  element upgrade time (scripts load at end of body).
- [Filtering by message text can be expensive for very large chats] → Session
  message counts are small in practice; direct iteration is acceptable. A
  debounce can be added later without changing the spec.
