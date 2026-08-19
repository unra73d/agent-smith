## 1. Title generation helper

- [x] 1.1 Add `Session.MaybeGenerateTitle(model *ai.Model)` in `src/agent/session.go`: guard on `s.temporary` and on the user-message count (0 or >3 → no-op).
- [x] 1.2 Implement message filtering (Origin == User or AI, non-empty trimmed Text) as an unexported helper used by `MaybeGenerateTitle`.
- [x] 1.3 Implement the title-generation goroutine: build filtered messages, call `ai.ChatCompletion` with the turn's model and a system prompt requesting a 2-4 word title, then sanitize the response — apply `util.CutThinking` first to strip any reasoning/thinking content, then trim whitespace/surrounding quotes.
- [x] 1.4 On success, set `s.Summary`, call `s.Save()`, and broadcast via the existing `SSEMessageSessionUpdate` pattern.
- [x] 1.5 Remove the naive `s.Summary = s.Messages[0].Text` assignment currently in `AddMessage`.

## 2. Wire into chat loops

- [x] 2.1 Call `session.MaybeGenerateTitle(model)` from `ToolChatStreaming`'s `AgentActionAnswer` case (`src/agent/tool_agent.go`), after `session.Save()`.
- [x] 2.2 Call `session.MaybeGenerateTitle(model)` from `DirectChatStreaming` (`src/agent/chat_agent.go`), after reply completion / before `streamDoneCh <- true`.

## 3. Tests

- [x] 3.1 Unit test: exchange-count eligibility (0, 1, 2, 3, 4+ user messages) gates regeneration correctly.
- [x] 3.2 Unit test: message filtering excludes tool-origin messages and empty/whitespace-only text.
- [x] 3.3 Unit test: title sanitization strips `<think>...</think>` (and equivalent) reasoning content, keeping only the model's final output.
- [x] 3.4 Unit test: temporary sessions never trigger title generation.
- [x] 3.5 Unit or integration test: successful title generation persists `Summary` and emits a `session_update` SSE event.

## 4. Verification

- [x] 4.1 `go build ./...`
- [x] 4.2 `go vet ./...`
- [x] 4.3 `go test ./...`
- [x] 4.4 `openspec validate kan-3 --strict`
