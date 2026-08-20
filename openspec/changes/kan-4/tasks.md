## 1. Chat find state and panel

- [x] 1.1 Add the top-right find panel markup and component state for query, debounce timer, match list, and active index.
- [x] 1.2 Implement Ctrl/Cmd+F opening, input focus, Escape/close dismissal, previous/next button behavior, and cancellation when a message is submitted.

## 2. Local matching and rendering integration

- [x] 2.1 Implement literal case-sensitive matching with a 500 ms debounce and muted/current match styles.
- [x] 2.2 Add safe text-node highlighting limited to visible user text and final model answers; exclude `.thinking-block` and `.tool-block` in every state, then add match count/empty state, navigation wraparound, and scroll-to-active behavior.
- [x] 2.3 Reapply or clear highlights after session changes and streamed assistant updates without backend calls; ensure message submission cancels all search state and pending work.

## 3. Verification

- [x] 3.1 Add or update frontend-focused validation where repository tooling permits, and manually verify keyboard shortcuts, case sensitivity, regex literals, navigation, closing, and session updates.
- [x] 3.2 Run `go build ./...`, `go vet ./...`, and `go test ./...`.
