.PHONY: help init_spec dev run server

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# AI SDLC targets
# ---------------------------------------------------------------------------

## Env required by init_spec (Confluence auth). Fail fast if missing.
REQUIRED_ENV := CONFLUENCE_URL JIRA_USERNAME JIRA_API_TOKEN CONFLUENCE_PARENT_ID

init_spec: ## Hydrate openspec/specs from Confluence (source of truth)
	@for v in $(REQUIRED_ENV); do \
		if [ -z "$$(printenv $$v)" ]; then echo "ERROR: env $$v is not set"; exit 1; fi; \
	done
	@command -v opencode >/dev/null 2>&1 || { echo "ERROR: opencode CLI not installed"; exit 1; }
	@test -f openspec/project.md || { echo "ERROR: OpenSpec not initialized. Run once and commit: openspec init --tools opencode --force"; exit 1; }
	docker pull ghcr.io/sooperset/mcp-atlassian:latest
	opencode run --agent spec-importer \
		"Import Confluence pages under parent ID $$CONFLUENCE_PARENT_ID into openspec/specs/ as OpenSpec spec files."

dev: ## Work a ticket locally with the interactive jira-dev agent (set JIRA_KEY)
	@test -n "$${JIRA_KEY}" || { echo "ERROR: set JIRA_KEY, e.g. make dev JIRA_KEY=KAN-1"; exit 1; }
	opencode run --agent jira-dev "Implement Jira ticket $${JIRA_KEY}"

# ---------------------------------------------------------------------------
# App targets
# ---------------------------------------------------------------------------

