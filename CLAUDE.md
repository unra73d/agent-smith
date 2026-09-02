# agent-smith — project instructions

Go desktop/server app (webview UI). Spec-driven development runs through
OpenSpec; tickets live in Jira; canonical specs live in Git under `openspec/specs/`.

## Entry points

- Select the OpenCode `implement` agent for the main local workflow. It reads the
  ticket, investigates relevant Jira and Confluence context, explores, proposes a
  spec delta, waits for approval, implements, tests, validates, syncs, archives,
  and then asks whether to commit to a feature branch.
- `/opsx:propose`, `/opsx:apply`, `/opsx:explore`, `/opsx:update`,
  `/opsx:archive`, `/opsx:sync` — raw OpenSpec operations, if you want to drive a
  single step by hand.

Subagents are dispatched by the workflow rather than invoked by hand: `coder`
(production implementation), `test-writer` (automated test coverage), and
`validator` (independent verification).

## Rules

- **`openspec/specs/` is the Git-tracked canonical baseline.** Spec changes begin
  as deltas under `openspec/changes/<KEY>/specs/`; after validation and explicit
  approval, run sync and archive to merge them into `openspec/specs/`.
- **Jira is read-only from this session.** Reading issues is fine; commenting,
  transitioning, assigning, and creating are blocked. Moving a ticket is the
  developer's call.
- **Never push or open a PR unprompted.** A local commit to a short, meaningful
  `feature/<jira-summary>` branch also requires explicit approval. Work on `main`
  is not allowed.
- **Never launch a nested OpenCode CLI** from an active agent session. Delegate
  with the Task tool instead.

## Checks

Keep the build green: `go build ./...`, `go vet ./...`, `go test ./...`.

## Credentials

The Atlassian MCP server (`@xuandev/atlassian-mcp`) is launched through
`@dotenvx/dotenvx`, which injects the gitignored `.env`. It needs
`ATLASSIAN_DOMAIN`, `ATLASSIAN_EMAIL`, and `ATLASSIAN_API_TOKEN`. See
`.env.example`; `make check` verifies them.
