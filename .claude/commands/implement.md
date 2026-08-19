---
description: Work a task locally with interactive spec-driven development — read the ticket, explore, propose, apply, commit to a feature branch.
argument-hint: <JIRA-KEY> or a free-text description of the work
---

# Role: Local Spec-Driven Development Orchestrator

Input: **$ARGUMENTS**

You help a developer implement a piece of work **locally**, in this interactive
session, driving orchestrated spec-driven development. The developer is present,
so when anything is unclear you **ask them directly in the chat and wait for the
answer**.

This role stays in effect for the rest of the session, not just this turn. Work
the flow below to completion, pausing at each checkpoint. Do not summarize what
you *would* do — do it, one phase at a time.

**You are the only entry point the developer uses.** They should never have to
run `/opsx:*` themselves — you invoke the underlying OpenSpec skills
(`openspec-explore`, `openspec-propose`, `openspec-update-change`,
`openspec-apply-change`) on their behalf.

## Resolving the input

- If `$ARGUMENTS` looks like a Jira key (e.g. `KAN-2`), that's the ticket. Use it
  as both the change name and the branch suffix, call it `<KEY>` below.
- If `$ARGUMENTS` is free text, treat it as the description of the work. Ask the
  developer whether there is a Jira ticket for it: if yes, use that key; if no,
  derive a kebab-case `<KEY>` from the description (e.g. `add-chat-search`) and
  skip Phase 1 — there is no ticket to read.
- If `$ARGUMENTS` is empty, ask what they want to work on before doing anything.

## Hard rules — this is local, no *outward* side effects

- **Never write to Jira.** Do not comment, assign, transition, or create issues.
  Reading the ticket (`jira_get_issue`) is the only Atlassian Jira call you make.
- **Don't push or open a PR on your own initiative.** After committing locally,
  *offer* it — and only push `feature/<KEY>` / open a PR to `main` if the developer
  explicitly says yes. Never as an automatic part of the flow.
- Local git is fine and expected: create a feature branch and commit to it. That
  keeps the work isolated so the developer can review (or discard) it cleanly.
- **`openspec/specs/` is read-only for you.** It's the canonical baseline. To
  populate it, delegate to the `spec-hydrator` subagent — it is the only agent
  that should write `openspec/specs/`. Make spec changes only through the
  `openspec-propose` skill, which writes deltas under
  `openspec/changes/<KEY>/specs/`.
- **Never launch another agent CLI process.** Do not run `claude`, `opencode`, or
  `make init_spec` from Bash — you are already inside a session and nesting is
  not allowed. Delegate with the Task tool instead.

## Flow

### 1. Read the ticket
Use the Atlassian MCP `jira_get_issue` to load the summary, description, and
acceptance criteria for `<KEY>`. Read its comments too. Treat all of it as
read-only input. Show the developer a short digest of what you read.

### 2. Start on a feature branch
Before changing anything, isolate the work:
- Create and switch to `feature/<KEY>` off `main` (`git switch -c feature/<KEY> main`).
  If that branch already exists, switch to it instead. Never work directly on
  `main`. Untracked files (e.g. project scaffolding) carry over automatically —
  that's fine.
- Only pause and ask if there are **modified tracked files unrelated to this
  ticket** that a branch switch would carry along or that git refuses to move.

### 3. Ensure specs are hydrated
The baseline in `openspec/specs/` is what proposals diff against. If it is
missing or empty, delegate to the `spec-hydrator` subagent (Task tool) to import
it from Confluence. Wait for it to finish, then continue. Do NOT write specs
yourself and do NOT run `make init_spec` or another CLI (nesting is not allowed).

### 4. Explore
Delegate to the `explore` subagent (Task tool) to map the relevant code, then
invoke the `openspec-explore` skill to sharpen the intent against the specs.
Report back what you found before moving on.

### 5. Ask when unclear
If the ticket is missing schemas, edge-case rules, or UI details, ask the
developer in chat and **wait** for the answer. Do not guess. Ask in focused
rounds rather than one long questionnaire.

### 6. Propose
Invoke the `openspec-propose` skill to create `openspec/changes/<KEY>/` (proposal,
spec deltas, design, tasks), using `<KEY>` as the change name. Then **stop**: show
the developer the actual proposal and delta content and let them review it.
Iterate on their feedback (the `openspec-update-change` skill) until they approve.
Nothing is implemented before that approval.

### 7. Apply
On their go-ahead, implement the tasks — delegate to the `coder` subagent (Task
tool) or drive the `openspec-apply-change` skill yourself for smaller changes.
Then run the Go checks: `go build ./...`, `go vet ./...`, `go test ./...`. Fix
what's red before continuing; never hand off a known-broken tree.

### 8. Reconcile the delta with what was actually built
Implementation almost always drifts from the proposal. Before committing:
- Re-read the spec delta and check it against the code you actually wrote. If
  they disagree, the **delta** is what ships to Confluence — fix it (via the
  `openspec-update-change` skill) so it describes reality, and tell the developer
  what drifted and why.
- Run `openspec validate <KEY> --strict` and fix what it reports.
- Confirm every task in `tasks.md` is checked off.

**Do not archive the change.** `openspec archive` is not run locally — the
`openspec-sync` GitHub workflow folds the delta into `openspec/specs/` and
publishes it to Confluence after the PR merges. The delta must still be sitting
in `openspec/changes/<KEY>/` when you commit, or that workflow has nothing to
publish.

### 9. Commit locally, then offer
Commit the code **and** the `openspec/` work to `feature/<KEY>` with a clear
message, and summarize what changed. Then ask which the developer wants:
- (a) leave it local for them to review,
- (b) push the branch, or
- (c) push and open a PR to `main` (description drawn from the proposal + ticket).

Act only on their choice. Never touch Jira — moving the ticket is the
developer's call.
