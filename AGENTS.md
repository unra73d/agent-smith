# agent-smith Project Guidelines

These instructions apply to all changes in this repository. Preserve the
hand-written, lightweight architecture and prefer direct, explicit code over
new frameworks or layers of abstraction.

## Architecture

- Keep package ownership clear:
  - `agent` owns application, chat, session, and role orchestration.
  - `ai` owns model/provider configuration and LLM communication.
  - `mcptools` owns MCP server lifecycle and tool integration.
  - `server` owns HTTP routes and SSE transport.
  - `ui` owns browser rendering and interaction.
- Keep transport, persistence, LLM behavior, MCP behavior, and browser rendering
  separate. Extend the existing owning package or component before adding a
  cross-cutting helper.
- Prefer small, direct implementations over frameworks, generic abstractions, or
  new dependencies. Use the standard library and existing project helpers first.
- Avoid unrelated refactors. Preserve public APIs, persisted data, SSE event
  compatibility, and existing behavior unless an approved OpenSpec change
  explicitly changes them.

## Preserve Architectural Intent

- Before changing a function signature, return type, state ownership, or
  communication path, determine the responsibility the existing design assigns
  to that boundary.
- Treat current data flow as intentional unless an approved requirement requires
  changing it. Trace where data is produced, persisted, transported, and
  consumed before adding another path for the same data.
- Prefer extending the established mechanism over adding a parallel one. Do not
  add return values, callbacks, shared state, or events when the information
  already has an intentional delivery path.
- Streaming and asynchronous functions may communicate through their supplied
  channels, callbacks, events, or owned state rather than return values. Preserve
  that contract unless a caller needs synchronous ownership of a result.
- Keep ownership singular: one component owns each concern, including session
  persistence, stream delivery, UI notification, and request lifecycle. Do not
  duplicate ownership just to make a local change easier.
- When a boundary's intent is unclear, inspect callers, neighboring functions,
  tests, and consumers before proposing a change. Ask the user when the evidence
  remains inconclusive.

## Error Handling

- Use the project logger's error helpers for operational failures instead of
  repetitive inline `if err != nil` blocks.
- Keep happy-path code linear. Use the logger helper to record failures and end
  the current operation when that is the established behavior.
- Do not introduce error-return plumbing, wrapper types, or custom recovery
  layers into existing flows unless a caller genuinely needs to make a
  domain-specific recovery decision.
- Return errors when the caller can recover, choose an alternative, or report a
  user-facing result. Otherwise follow the local logger pattern.
- Preserve contextual logging: errors must identify the failed operation and the
  relevant resource or identifier.

## Go and Concurrency

- Follow established Go style: standard library first, explicit control flow,
  small functions, and focused files.
- Keep HTTP handlers thin. Put stateful behavior in `agent`, LLM/provider logic
  in `ai`, and MCP logic in `mcptools`.
- Prefer small, explicit Go channels for completion signals, streaming results,
  and coordination between a request handler and background work.
- Keep channels single-purpose and owned by the goroutine that creates them.
- Prefer a simple completion channel over `sync.WaitGroup`, callbacks, futures,
  worker pools, or extra abstraction when one producer and one consumer are
  sufficient.
- Use goroutines only for real asynchronous work. Ensure each completion path
  signals, closes, or otherwise releases its waiting counterpart.
- Maintain existing Gin route grouping and JSON request/response shapes.
- Keep SQLite persistence local to its owning domain and serialize structured
  persisted fields deliberately.

## UI and Events

- Keep the UI framework-free: browser APIs, Web Components, Shadow DOM, plain
  JavaScript, and existing CSS patterns.
- Use SSE to publish backend state changes and document-level custom events to
  distribute them in the browser.
- Keep cross-component state event-driven through the existing `Storage` object
  and event mechanism. Do not introduce a UI framework, controller hierarchy, or
  class-based state layer.
- Add a typed server event and corresponding client handler when a backend state
  change must reach the UI.
- Keep Web Components focused on rendering and local interaction; use shared
  events only for cross-component state.
- Add browser API wrappers in `src/ui/api.js` using the existing `api<Action>`
  naming and error-handling pattern.
- Keep component state local unless another component genuinely needs it.
- Put global primitives in `src/ui/global.css`, application layout and shared
  styling in `src/ui/styles.css`, and component-specific styling with its
  component.

## Testing and Change Discipline

- Every behavior change needs automated coverage at the closest practical level.
- Use existing same-package `*_test.go` patterns and established temporary
  database and `httptest` helpers.
- Do not weaken, delete, or skip tests to make a change pass.
- For Go changes, run `go test ./...`, `go vet ./...`, and `go build ./...`.
- Read related code and tests before editing. State unresolved behavior and
  assumptions before implementation.
- Use the OpenSpec workflow for behavior changes: explore, propose, user
  approval, implementation, automated tests, validation, sync, and archive.
- In a proposal, state any architectural boundary being changed, why the existing
  path is insufficient, and the compatibility impact.
- Work on a concise `feature/<jira-summary>` branch. Never commit, push, open a
  pull request, or change Jira without explicit user approval.
