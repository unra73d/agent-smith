## Context

See proposal.md - Why. Relevant existing code:

- `Session.Summary` (`src/agent/session.go`) already exists and is already persisted (`sessions.summary` DB column) and broadcast (`session_update` SSE carries the whole `*Session`). `AddMessage` currently sets it once, naively, from the raw first user message.
- Two independent chat loops reach "assistant reply done" differently: `ToolChatStreaming` (`src/agent/tool_agent.go`) at the `AgentActionAnswer` case, and `DirectChatStreaming` (`src/agent/chat_agent.go`) right before signaling `streamDoneCh`.
- `MessageOrigin` (`src/ai/message.go`) distinguishes `User`, `AI`, `Tool`, `System`.
- `ai.ChatCompletion(messages, sysPrompt, model, tools)` (`src/ai/apiprovider.go`) is the existing synchronous call used for one-off model prompts (see `DynamicAgentChat` in `tool_agent.go` for the established calling convention: resolve model via `findModel(modelID)`, build a message slice, call `ChatCompletion`).

## Goals / Non-Goals

**Goals:**
- One shared code path for title (re)generation, called from both chat loops, so the exchange-counting and message-filtering logic isn't duplicated.
- No new persisted fields or DB migrations.
- No frontend changes.

**Non-Goals:**
- Not introducing a separate/cheaper "titling model" concept — out of scope per proposal.
- Not changing how exchanges are counted if a user deletes/truncates messages — the count is a live derived value (see Decisions).

## Decisions

**Shared helper on Session.** Add a method, e.g. `Session.MaybeGenerateTitle(model *ai.Model)`, called from both `ToolChatStreaming`'s `AgentActionAnswer` case and `DirectChatStreaming`'s reply-completion point, right after each already calls `session.Save()`. It:
1. Returns immediately if `s.temporary` is true.
2. Counts `Origin == MessageOriginUser` messages in `s.Messages`; returns immediately if the count is 0 or > 3.
3. Launches a goroutine that:
   - Builds a filtered `[]*ai.Message` containing only messages with `Origin == MessageOriginUser` or `Origin == MessageOriginAI` and non-empty (post-trim) `Text`.
   - Calls `ai.ChatCompletion` with a short system prompt instructing a 2-4 word title, using the same `model` the chat turn used.
   - Sanitizes the response — `ChatCompletion` returns raw model content with no thinking-tag stripping (unlike the streaming path, which applies `util.CutThinking` at `apiprovider.go:519`), so the helper must call `util.CutThinking` on the result first, then trim whitespace/surrounding quotes — and sets `s.Summary` from what remains after the closing `</think>` (or equivalent) tag.
   - Persists via `s.Save()` and broadcasts via the same `SSEMessageSessionUpdate` pattern already used elsewhere in `session.go`, mirroring how `AddMessage`/`UpdateLastMessage` push updates.

Why a method on `Session` rather than a free function in `agent` package: `Session` already owns `Summary`, `Messages`, `Save()`, and the SSE-push convention: keeping this colocated avoids a new file's worth of indirection for ~30-40 lines of logic. Alternative considered: a package-level `agent.GenerateSessionTitleIfNeeded(session, model)` — rejected only because it would need the same private fields/methods `Session` already exposes to itself, adding no isolation benefit.

**Exchange counting is a live count, not a stored counter.** Counting `Origin == MessageOriginUser` messages in `s.Messages` on each call is O(n) in message count, but n is small (at most a handful of messages by definition — this stops running after the 3rd user message) and this rides on data already loaded in memory for the request. Alternative considered: a persisted `exchangeCount int` field — rejected because it adds a schema/model field purely to avoid an O(n) scan that's bounded to n≤~6 messages by the requirement itself, and a live count naturally self-corrects if messages are pruned.

**Fire-and-forget goroutine, not synchronous.** The title call is a second `ChatCompletion` invocation (with its own `WaitForAllowance()` rate-limit wait) on top of the turn's main completion; running it inline would add visible latency to every one of the first 3 replies. Since `session_update` is already a live-push mechanism independent of the HTTP response cycle, firing the title update slightly after the reply is consistent with how the UI already consumes session changes.

**Reuse the turn's own model.** No new model-selection concept is introduced; the same `modelID` already passed into `ToolChatStreaming`/`DirectChatStreaming` is resolved via the existing `findModel` and reused for the title call, matching `DynamicAgentChat`'s established pattern of reusing the caller-specified model for auxiliary calls.

## Risks / Trade-offs

- **Extra model calls / cost** → bounded to at most 3 short calls per session (title generation stops after exchange 3), and reuses the already-budgeted `WaitForAllowance()` rate limiter so it can't bypass existing throttling.
- **Race: two goroutines regenerating the title concurrently** if a user fires messages faster than titles resolve → out of scope to fully solve here; `Session.Save()` already serializes on whatever locking `Session` currently uses for message mutation, and the last write wins, which is an acceptable outcome for a title field (not correctness-critical state).
- **Model returns something longer than 2-4 words or wraps it in quotes/punctuation** → mitigated by a small sanitize/trim step in the helper; not a hard guarantee, since enforcing an exact word count from a free-text model response isn't reliable — the system prompt asks for it, sanitization cleans obvious formatting artifacts, but there is no hard truncation to avoid mangling a legitimate short phrase.
- **Reasoning models emit `<think>...</think>` (or similar) blocks before the actual title** → `ChatCompletion` does not strip these (unlike the streaming path). The helper applies `util.CutThinking` to the raw response before any other sanitization, so `Summary` is set from the model's final output only, never from reasoning content.
