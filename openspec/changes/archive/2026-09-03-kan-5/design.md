## Context

See proposal.md for motivation. Streaming providers currently deliver text through the `ai` package to an `agent.Session`, which emits existing SSE message updates consumed by the UI. The provider parses terminal stream chunks but does not retain usage metadata; `ai.Message` is serialized as the session JSON blob. No separate response route or SSE event is needed.

## Goals / Non-Goals

**Goals:**
- Capture a successful final answer's output-token count and elapsed generation duration.
- Use the existing `ai -> agent session -> SSE -> Storage -> ChatView` ownership path.
- Preserve optional statistics across persisted and legacy sessions.

**Non-Goals:**
- Per-session aggregates, prompt-token/cost reporting, time-to-first-token, or provider/model attribution.
- Statistics for dynamic-agent calls, tool-result messages, intermediate tool-call assistant turns, failed responses, or cancelled responses.
- A new database table, API endpoint, SSE event type, UI framework, or dependency.

## Decisions

### Optional message statistics are the persisted contract

Add optional output-token and elapsed-milliseconds fields to the assistant message representation. SQLite already persists messages as a JSON blob, so this extends the owned session representation without a schema migration. Zero/absent fields mean statistics are unavailable and preserve compatibility with stored historical messages.

Alternative considered: a separate response-statistics table. Rejected because it duplicates message ownership and would require correlation, migration, and a second persistence path.

### Provider usage is advisory; local estimation is the fallback

The OpenAI-compatible streaming response parser retains completion-token usage when terminal metadata provides it and exposes it to the caller only on successful completion. The agent estimates output tokens from final assistant text only when usage is absent. Providers that do not send usage remain supported; no provider-specific request option is required.

Alternative considered: require `stream_options.include_usage`. Rejected because it is not universally supported and could break configured OpenAI-compatible providers.

### Agent records only the final successful answer

Each provider invocation measures elapsed time at its orchestration boundary. Direct chat attaches statistics after a successful stream. Tool chat retains each invocation result while looping, but attaches statistics only once the final answer classification succeeds; tool-call turns remain unannotated. Failed/cancelled streams never attach statistics. The final assignment uses the existing last-message update and session save path, avoiding a parallel client transport.

Alternative considered: annotate every provider invocation. Rejected by the approved scope because intermediate tool turns are not visible final answers.

### UI footer is separate from markdown content

`ChatView` renders a dedicated footer outside `.message-content` so streaming markdown rerenders cannot erase it. The UI retains the optional fields from both full session and live message event payloads, recalculates the footer on the terminal last-message update, and hides it when fields are absent or non-positive. Valid statistics reserve their fixed footer space and use visibility rather than display to reveal on the existing message-level hover selector at the same time as message action buttons, preventing chat-layout movement.

Alternative considered: append statistics to assistant text. Rejected because it would pollute persisted model content, copying, markdown/search behavior, and tool parsing.

## Risks / Trade-offs

- Provider usage often is not present in streams -> deterministic local estimation provides a consistent display, but is approximate.
- Very short durations can create unstable rates -> use a minimum one-second divisor and whole-number formatting.
- Asynchronous stream consumers can race final assignment -> ensure all text updates for a completed invocation are applied before estimating/attaching final statistics.
- Tool-loop state can associate a result with the wrong message -> attach only after action classification confirms the current turn is the successful final answer.

## Migration Plan

Deploy as a backward-compatible message JSON extension. Existing sessions load with no statistics and render unchanged. Rollback consists of reverting consumers; optional JSON fields are ignored by older binaries.

## Test Strategy

### `src/ai/apiprovider_test.go`: `TestChatCompletionStream_CapturesTerminalCompletionUsage`
- **Setup/inputs/mocks:** An `httptest` OpenAI-compatible SSE endpoint streams content chunks, a terminal empty-choice usage chunk, and `[DONE]`.
- **Action:** Run `ChatCompletionStream` with buffered text consumption.
- **Expected observable result:** Text deltas are forwarded and the successful result exposes the terminal completion-token count.
- **Requirement/scenario:** ai-providers “Stream text and usage”; ai-message-statistics “Provider usage is available.”

### `src/ai/apiprovider_test.go`: `TestChatCompletionStream_ReportsMissingUsage`
- **Setup/inputs/mocks:** An `httptest` SSE endpoint streams text and `[DONE]` without usage metadata.
- **Action:** Run `ChatCompletionStream`.
- **Expected observable result:** The stream succeeds and reports unavailable provider usage.
- **Requirement/scenario:** ai-providers “Stream omits usage”; ai-message-statistics “Provider usage is unavailable.”

### `src/ai/stats_test.go`: `TestEstimateOutputTokensAndFormatStatistics`
- **Setup/inputs/mocks:** Empty, short, normal, and long generated text plus zero, sub-second, seconds, and minutes-plus-seconds durations.
- **Action:** Estimate tokens and format throughput/duration.
- **Expected observable result:** Empty text yields no usable statistics; positive output produces a whole-number `t/s` rate with guarded duration formatting.
- **Requirement/scenario:** ai-message-statistics “Provider usage is unavailable” and “Completed answer has statistics.”

### `src/agent/session_test.go`: `TestResponseStatisticsPersistAndBroadcastFinalMessage`
- **Setup/inputs/mocks:** A temporary SQLite session with an assistant final answer and an SSE receiver.
- **Action:** Record valid response statistics and reload the session.
- **Expected observable result:** The existing `last_message_update` contains the expanded assistant message; fields survive session JSON persistence.
- **Requirement/scenario:** sessions “Completed answer data” and “Final statistics update.”

### `src/agent/session_test.go`: `TestLegacyMessageLoadsWithoutResponseStatistics`
- **Setup/inputs/mocks:** A persisted legacy message JSON payload without optional fields.
- **Action:** Load the session.
- **Expected observable result:** Loading succeeds and the message exposes no usable statistics.
- **Requirement/scenario:** sessions “Legacy message data.”

### `src/agent/chat_agent_test.go`: `TestDirectChatRecordsStatisticsOnlyAfterSuccessfulCompletion`
- **Setup/inputs/mocks:** A controllable streaming provider returns either success or cancellation/failure for a direct chat.
- **Action:** Run direct chat to its completion signal.
- **Expected observable result:** The successful assistant answer has statistics; failed/cancelled assistant messages do not.
- **Requirement/scenario:** agent-orchestration “Successful direct answer” and “Failed or cancelled direct answer.”

### `src/agent/tool_agent_test.go`: `TestToolChatRecordsStatisticsOnlyOnFinalAnswer`
- **Setup/inputs/mocks:** A controllable provider first requests a tool and then returns a successful final answer; the tool returns a deterministic result.
- **Action:** Run tool chat to completion.
- **Expected observable result:** The intermediate assistant tool-call message has no statistics and the final assistant answer has them.
- **Requirement/scenario:** agent-orchestration “Tool-assisted final answer.”

### UI verification
- **Automated-test status:** The repository has no existing browser/JavaScript test harness. Adding one solely for this small rendering update would introduce a new dependency and testing architecture outside the approved scope.
- **Focused verification:** Inspect the browser UI with a final assistant message containing valid fields and one without them; confirm the footer is muted, beneath markdown, reserves its space without shifting nearby messages, appears only while its message is hovered alongside action buttons, updates after SSE, and is omitted when unavailable.
- **Requirement/scenario:** web-ui “Render completed answer statistics,” “Render an answer without statistics,” and “Final response-statistics update.”
