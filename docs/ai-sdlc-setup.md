# AI SDLC — Autonomous Virtual Worker Setup

This repo implements the spec-driven autonomous worker described in [`sdd.md`](../sdd.md),
built on **OpenCode** (headless AI runtime) + **OpenSpec** (spec engine), with
**mcp-atlassian** bridging Jira and Confluence.

Flow: Jira transition → GitHub `repository_dispatch` → Actions runner → OpenCode
(`jira-worker` agent) → OpenSpec propose/apply → PR → human merge → post-merge
archive + Confluence sync.

## What's already in the repo

| File | Purpose |
|------|---------|
| `.opencode/opencode.json` | OpenCode model + Atlassian MCP (Docker) config |
| `.opencode/agents/jira-worker.md` | **CI** orchestrator — autonomous, Jira comments, opens PR |
| `.opencode/agents/jira-dev.md` | **Local** orchestrator — interactive, asks you, no Jira writes / no PR |
| `.opencode/agents/explore.md` | Read-only exploration subagent |
| `.opencode/agents/coder.md` | Implementation subagent |
| `.opencode/agents/spec-importer.md` | Confluence → OpenSpec importer agent |
| `.github/workflows/virtual-worker.yml` | Webhook-triggered CI execution pipeline |
| `.github/workflows/openspec-sync.yml` | Post-merge archive + scoped Confluence sync |
| `Makefile` | `make init_spec` hydrates specs from Confluence; `run`/`server` |

## Two ways to run the pipeline

Same OpenSpec flow, two contexts — no versioning, feature branches cut from `main`:

- **CI (`jira-worker`)** — headless, no human present. Communicates through Jira:
  comments clarifications, reassigns on ambiguity, opens the PR. Triggered by the
  Jira → `repository_dispatch` webhook.
- **Local (`jira-dev`)** — interactive. Reads the ticket as input, hydrates specs,
  runs `/opsx:propose` + `/opsx:apply`, and **asks you in the chat** when unclear.
  Never writes to Jira, never pushes or opens a PR — you drive git.

## Spec lifecycle & git tracking

**Confluence is the permanent home for specs. `main` holds code only — no
`openspec/`.** Specs exist in git only transiently, on feature branches.

1. Cut `feature/JIRA-123` off `main` (code only).
2. `make init_spec` downloads the current specs from Confluence into
   `openspec/specs/` (the base `openspec archive` folds deltas onto).
3. `/opsx:propose` + `/opsx:apply` run; the feature branch commits code + the whole
   `openspec/` tree (hydrated specs + `openspec/changes/JIRA-123/` deltas) so the
   PR shows the spec work for review.
4. **On merge to `main`**, `openspec-sync.yml`: runs `openspec archive` to fold the
   deltas into the specs → publishes **only the spec pages that changed** to
   Confluence (diffed from what `archive` touched) → `rm -rf openspec/` and commits
   the deletion, so only code lands on `main`.

> Note: the archive base is whatever Confluence held when the feature branch was
> hydrated. If two tickets touch the same capability concurrently, the second
> merge publishes over the first — serialize such tickets, or add a re-hydrate +
> conflict check if that becomes real.

## Setup — steps only you can do

### 1. Seed Confluence with an initial spec set (only if starting empty)
CI hydrates specs from Confluence on every run, so Confluence must contain them.
If you're adopting this on an existing system with no specs yet, generate a first
set and publish them to Confluence once:
```bash
export CONFLUENCE_URL=... JIRA_USERNAME=... JIRA_API_TOKEN=... CONFLUENCE_PARENT_ID=...
make init_spec        # or `openspec init` + write specs by hand
# then publish openspec/specs/ to Confluence (see the sync workflow's action)
```
If Confluence already documents the system, skip this — the importer reads it.

### 2. Rotate the exposed token, then store secrets
The token in `tokens.md` was shared in plaintext (now gitignored). If this repo
is or will be public, **rotate it** at https://id.atlassian.com/manage-profile/security/api-tokens.
Then add these under **Settings → Secrets and variables → Actions**:

| Secret | Value |
|--------|-------|
| `OPENCODE_API_KEY` | API key for your `go` OpenCode provider (rename to match its expected env var) |
| `JIRA_URL` | `https://<your-domain>.atlassian.net` |
| `JIRA_USERNAME` | your Atlassian account email |
| `JIRA_API_TOKEN` | Atlassian API token (the one in `tokens.md`, rotated) |
| `CONFLUENCE_URL` | `https://<your-domain>.atlassian.net/wiki` |
| `CONFLUENCE_PARENT_ID` | root page ID for published specs (from the page URL) |
| `BOT_GITHUB_TOKEN` | PAT (or fine-grained token) with repo + PR scope |

> The same `JIRA_API_TOKEN` / `JIRA_USERNAME` are reused for Confluence auth by
> both the MCP server and the publish action.

### 3. Trigger: use a Jira Automation rule (not the legacy webhook)
The SDD's webhook JQL `status CHANGED TO "In Progress"` can't be expressed in a
webhook scope filter. Use **Project settings → Automation → Create rule** instead:

- **Trigger:** *Issue transitioned* → To status: `In Progress`
- **Condition:** *Issue fields condition* → Assignee = the virtual-worker account
- **Action:** *Send web request*
  - URL: `https://api.github.com/repos/<OWNER>/<REPO>/dispatches`
  - Method: `POST`
  - Headers: `Authorization: Bearer <BOT_GITHUB_TOKEN>`, `Accept: application/vnd.github+json`
  - Body:
    ```json
    { "event_type": "jira-task-assigned", "client_payload": { "jira_key": "{{issue.key}}" } }
    ```

### 4. Confluence publishing (already wired — nothing to configure)
`openspec-sync.yml` injects the required frontmatter automatically: for each
changed `spec.md` it derives a unique `connie-title` from the capability folder
name and sets `connie-publish: true`, then publishes only those staged files
under `CONFLUENCE_PARENT_ID`. Pages are matched by title under the parent, so the
first publish of a capability creates its page and later publishes update it. No
`.markdown-confluence.json` or per-file frontmatter is needed.

**Confluence page titles = capability folder names.** `openspec/specs/user-auth/`
publishes as page **"User Auth Spec"**. Keep capability slugs stable so pages
update in place instead of forking a new page on rename.

## Corrections applied vs. the original sdd.md

- MCP: `@modelcontextprotocol/server-jira` (nonexistent) → `mcp-atlassian` (Jira + Confluence, API-token auth, headless-friendly).
- Dropped the deprecated `@modelcontextprotocol/server-github` MCP — PRs are created with the `gh` CLI.
- opencode.json: `"type": "stdio"` → `"type": "local"`; literal `ENV_*` values → `{env:...}` interpolation.
- Model `claude-3-7-sonnet` → `opencode-go/deepseek-v4-flash` (the `go` OpenCode provider); provider key secret `OPENCODE_API_KEY`.
- Confluence publish scoped to only the spec pages `archive` changed (was: whole-folder republish).
- Specs never persist in git: hydrated from Confluence per feature branch, stripped from `main` on merge.
- Jira trigger: legacy webhook JQL → Automation rule on issue-transitioned.
- Removed release/version management — feature branches cut straight from `main`.
- Split into CI (`jira-worker`, autonomous) and local (`jira-dev`, interactive, no side effects) agents.
- Added Go build/test steps and `[skip ci]` loop guards.

## Local development

Provider auth is stored by `opencode auth login` (not an env var locally); only the
Atlassian MCP needs shell env. `opencode-go` is the provider; docker must be running.
Use the **`jira-dev`** agent — it reads the ticket, asks you when unclear, and makes
no Jira/PR side effects. (Do **not** run `jira-worker` locally; that's the CI agent
and it comments on Jira and opens PRs.)

```bash
# One-time: authenticate the provider and make sure docker is up
opencode auth login                # pick the opencode-go provider
docker pull ghcr.io/sooperset/mcp-atlassian:latest

# Load Atlassian env into your shell (opencode reads process env — no .env autoload)
cp .env.example .env               # then fill in JIRA_* / CONFLUENCE_URL
set -a; . ./.env; set +a

# Hydrate the canonical specs from Confluence (baseline for OpenSpec)
make init_spec

# Work a ticket interactively — asks you when the ticket is ambiguous
opencode                           # then, in the TUI, switch to the `jira-dev` agent
# or one-shot:
opencode run --agent jira-dev "Implement Jira ticket KAN-1"
```

> For a quick connectivity check before anything else, read-only:
> `opencode run "Fetch Jira issue KAN-1 via the atlassian MCP and print its summary."`
