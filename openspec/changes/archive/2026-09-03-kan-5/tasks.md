## 1. Provider and message statistics contract

- [x] 1.1 In `src/ai/message.go` and a focused `src/ai` statistics helper, add optional output-token and elapsed-duration message fields plus deterministic estimation and compact formatting helpers; verify unset fields remain absent from serialized legacy-compatible message JSON. (ai-message-statistics; sessions Message model)
- [x] 1.2 In `src/ai/apiprovider.go`, retain completion-token usage from terminal OpenAI-compatible SSE metadata and expose it only for a successfully completed stream while preserving delta and tool-call behavior; verify streams without usage remain successful. (ai-providers Streaming chat completion)

## 2. Final-answer orchestration and session delivery

- [x] 2.1 In `src/agent/session.go`, add the owned operation that records valid response statistics on the current final assistant message, persists non-temporary sessions, and emits the existing `last_message_update`; verify its payload includes the expanded message. (sessions Streaming last-message update)
- [x] 2.2 In `src/agent/chat_agent.go`, measure a direct-chat invocation and record statistics only after its successful completion and final text consumption; verify failed and cancelled direct streams leave statistics absent. (agent-orchestration Direct chat session semantics)
- [x] 2.3 In `src/agent/tool_agent.go`, retain per-invocation completion results but record statistics only after the tool loop classifies the current assistant message as its successful final answer; verify intermediate tool-call turns remain unannotated. (agent-orchestration Answer termination)

## 3. Chat UI rendering

- [x] 3.1 In `src/ui/api.js` and `src/ui/app.js`, retain optional response-statistics fields from new-message, last-message-update, and full-session payloads; verify the existing event path supplies final message data to `ChatView`. (web-ui Real-time message updates)
- [x] 3.2 In `src/ui/components/chat/chat.js` and `chat.css`, render a dedicated compact muted footer beneath assistant markdown, calculate whole-number throughput and compact duration, hide it for unavailable fields, reserve its space so hover visibility does not shift surrounding messages, and reveal valid statistics only while the containing assistant message is hovered alongside its action buttons; verify markdown rerenders do not remove the footer. (ai-message-statistics Display final-answer response statistics; web-ui Message rendering)

## 4. Automated coverage

- [x] 4.1 Add `TestChatCompletionStream_CapturesTerminalCompletionUsage` in `src/ai/apiprovider_test.go` using an `httptest` SSE provider; assert forwarded text and returned terminal completion tokens. (ai-providers Stream text and usage)
- [x] 4.2 Add `TestChatCompletionStream_ReportsMissingUsage` in `src/ai/apiprovider_test.go` using an `httptest` SSE provider without usage metadata; assert successful completion and unavailable usage. (ai-providers Stream omits usage)
- [x] 4.3 Add `TestEstimateOutputTokensAndFormatStatistics` in `src/ai/stats_test.go`; assert estimation, whole-number throughput, zero-duration guard, and compact seconds/minutes formatting. (ai-message-statistics Provider usage is unavailable; Completed answer has statistics)
- [x] 4.4 Add `TestResponseStatisticsPersistAndBroadcastFinalMessage` in `src/agent/session_test.go` with the existing temporary SQLite/SSE helpers; assert the final update event and reloaded session retain statistics. (sessions Completed answer data; Final statistics update)
- [x] 4.5 Add `TestLegacyMessageLoadsWithoutResponseStatistics` in `src/agent/session_test.go`; assert a legacy message JSON payload loads successfully with no usable statistics. (sessions Legacy message data)
- [x] 4.6 Add `TestDirectChatRecordsStatisticsOnlyAfterSuccessfulCompletion` in `src/agent/chat_agent_test.go` with a controllable stream outcome; assert success records statistics and failure/cancellation does not. (agent-orchestration Successful direct answer; Failed or cancelled direct answer)
- [x] 4.7 Add `TestToolChatRecordsStatisticsOnlyOnFinalAnswer` in `src/agent/tool_agent_test.go` with a deterministic tool-call then answer flow; assert only the final answer has statistics. (agent-orchestration Tool-assisted final answer)

## 5. Validation

- [ ] 5.1 Run focused `go test ./src/ai ./src/agent` coverage and inspect a completed and an unavailable-statistics assistant message in the UI; verify the muted footer is below markdown, reserves space without shifting nearby messages, appears only on its containing message hover alongside action buttons, updates through SSE, and is omitted when absent. (all response-statistics scenarios)
  - Evidence: focused Go tests pass. Implementation inspection confirms the footer is a sibling after `.message-content`, `last_message_update` merges the complete message before `ChatView` rerenders markdown then statistics, and unavailable/non-positive fields set the footer hidden. Runtime server startup served `/ui/`, but configured unavailable MCP/Ollama services prevented a browser chat-flow inspection; 5.1 remains unmarked.
- [ ] 5.2 Run `go test ./...` and verify the complete repository test suite passes.
- [ ] 5.3 Run `go vet ./...` and verify static analysis passes.
- [ ] 5.4 Run `go build ./...` and verify the application builds.
- [ ] 5.5 Run `openspec validate kan-5 --strict` and verify the proposal, design, tasks, and delta specifications are coherent.
