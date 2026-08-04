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

- **Style the search field from the main document stylesheet, not the shadow
  root.** Because the field markup is in the light DOM, its styles must live in
  `src/ui/styles.css` (loaded via `<link>` in `index.html`). The
  `sessions.css` stylesheet is adopted only into the `<session-list>` shadow
  root and would not reach the field — initially the styles were placed there,
  which made the input render with the default white browser style and let the
  clear button wrap to the next line (review feedback). The fix moved the
  `.session-search*` rules into `styles.css`.

- **Match against the displayed title and raw message text.**
  The list shows `session.summary` (or "New chat" when empty), so matching uses
  the same fallback. Message matching iterates `session.messages` and checks
  `msg.text`. Case-insensitivity via `.toLowerCase()` on both sides.

- **Reuse existing CSS patterns.**
  Colors, borders, and the `img-button` icon styling follow the existing dark
  theme (`#313335` input backgrounds, `#3c4043` borders, `#8ab4f8` focus) — the
  same palette used by the `.filter-input` and the dropdown `<select>`, so the
  search field is visually consistent with the app's other text inputs.

## Risks / Trade-offs

- [Search field styling lives in the wrong document (shadow root vs light DOM)]
  → The search field markup is in the light DOM, so its styles must be in
  `src/ui/styles.css`. Placing them in `sessions.css` (adopted into the
  `session-list` shadow root) leaves the field unstyled — white input and the
  clear button on its own line. This was caught in review and fixed; the note in
  `sessions.css` guards against regression.
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
