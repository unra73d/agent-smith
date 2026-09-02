# AI SDLC Setup

This repository uses an interactive, spec-driven delivery workflow built on
OpenCode, OpenSpec, and the Atlassian MCP. Select the `implement` agent to work
a Jira ticket with a developer present.

There is no unattended Jira worker, Jira automation rule, automatic Confluence
publishing, automatic push, or pull-request creation. Jira and Confluence are
read-only sources of context. Canonical OpenSpec files are committed under
`openspec/specs/`.

## Prerequisites

Install and authenticate the local tools:

```bash
npm install -g opencode-ai @fission-ai/openspec
opencode auth login
```

The repository uses the `opencode-go/deepseek-v4-flash` provider. Select that
provider during `opencode auth login`.

Create the local Atlassian MCP environment file:

```bash
make init_env
```

Set these values in `.env`:

```dotenv
ATLASSIAN_DOMAIN=your-site.atlassian.net
ATLASSIAN_EMAIL=you@example.com
ATLASSIAN_API_TOKEN=your-api-token
```

Create the token at
<https://id.atlassian.com/manage-profile/security/api-tokens>. The file is
gitignored and is loaded automatically by Make targets.

Check the local setup:

```bash
make check
```

## Work A Ticket

Start the workflow with a Jira key:

```bash
make dev JIRA_KEY=KAN-1
```

Or start OpenCode directly after loading `.env` into the shell, then select the
`implement` agent and provide the Jira key:

```bash
set -a; . ./.env; set +a
opencode
```

The `implement` agent follows this sequence:

1. Reads the ticket, then creates or switches to a concise
   `feature/<short-kebab-summary>` branch derived from its Jira summary, without
   discarding unrelated local work.
2. Uses the `ticket-system` skill to read the ticket, comments, and relevant
   linked Jira items. It uses the `knowledgebase` skill to read relevant
   Confluence pages and related pages when they clarify the change.
3. Explores the repository and existing OpenSpec files, then asks the developer
   about any material ambiguity.
4. Runs OpenSpec exploration and proposes a change under
   `openspec/changes/<JIRA-KEY>/`. The developer reviews and explicitly approves
   the proposal before implementation begins.
5. Delegates production changes to `coder`, automated test coverage to
   `test-writer`, and independent checks to `validator`.
6. When validation fails, returns the actionable findings to `coder` and repeats
   implementation, test writing, and validation. It pauses for developer
   direction after five repair cycles.
7. When validation passes, presents the evidence and asks the developer to
   approve finalization.
8. After approval, syncs the delta into `openspec/specs/`, then archives the
   completed change. Canonical specs remain tracked in Git.
9. Asks whether to create a local commit with code, tests, and canonical specs.
   It never commits, pushes, opens a pull request, or changes Jira without a
   separate explicit request.

## Agent Responsibilities

`implement` is the sole primary agent. It owns user interaction, research,
approvals, delegation, retry control, and finalization.

- `coder` implements approved production tasks only.
- `test-writer` adds focused automated Go tests from approved requirements and
  implementation changes.
- `validator` independently maps requirements to evidence and runs `go test
  ./...`, `go vet ./...`, `go build ./...`, and applicable OpenSpec verification.

The `ticket-system` and `knowledgebase` skills selectively follow Jira comments,
Jira links, and Confluence links when they materially clarify requirements,
dependencies, contracts, or prior decisions. They do not crawl unrelated items
and never write to Atlassian.

## OpenSpec Lifecycle

`openspec/specs/` is the version-controlled canonical baseline. Work starts as
a proposal and delta specs under `openspec/changes/<JIRA-KEY>/`. The primary
agent syncs and archives only after the proposal is approved, implementation is
validated, and the developer approves finalization.

You can use the generated OpenSpec skills directly for a single step:

- `/opsx-explore`
- `/opsx-propose`
- `/opsx-update`
- `/opsx-apply`
- `/opsx-verify`
- `/opsx-sync`
- `/opsx-archive`

Use the `implement` agent for the full workflow; raw operations do not replace
its approval and validation gates.
