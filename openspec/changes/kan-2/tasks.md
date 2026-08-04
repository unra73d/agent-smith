## 1. Markup and Styles

- [x] 1.1 Add the search field wrapper, input, and clear button to the Chat sessions tab in `src/ui/index.html`
- [x] 1.2 Style the search field and clear button in `src/ui/styles.css` (the main document stylesheet — the field markup lives in the light DOM, so shadow-root styles in `sessions.css` would never reach it)
- [x] 1.3 Adjust the `.session-list` max-height in `src/ui/styles.css` so the list scrolls with the search bar pinned above it
- [x] 1.4 Apply the review fix: style the field with the app's dark theme (consistent with `.filter-input` / dropdown inputs) and keep the clear button absolutely positioned inside the field so it sits on the same line as the input

## 2. Filtering Logic

- [x] 2.1 Add filter state and wire the search input + clear button events in `src/ui/components/sessions/sessions.js`
- [x] 2.2 Override `updateList()` in `SessionList` to render only sessions matching the current filter
- [x] 2.3 Implement case-insensitive matching against the chat title (summary) and message text

## 3. Verification

- [x] 3.1 Run `go build ./...`, `go vet ./...`, `go test ./...` and confirm the repo stays green
