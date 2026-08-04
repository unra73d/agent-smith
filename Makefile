.PHONY: help init_env init_spec dev tui run server

# Auto-load .env (KEY=VALUE) and export to every recipe, if the file exists.
# This is why you don't need `set -a; . ./.env; set +a` before make targets.
-include .env
export JIRA_URL JIRA_USERNAME JIRA_API_TOKEN CONFLUENCE_URL CONFLUENCE_PARENT_ID OPENCODE_API_KEY

help: ## Show available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

init_env: ## Create .env from .env.example (fill it in afterwards)
	@if [ -f .env ]; then \
		echo ".env already exists — leaving it untouched."; \
	else \
		cp .env.example .env && echo "Created .env from .env.example — fill in your values."; \
	fi
	@echo "Make targets auto-load .env. To load it into your own shell for running"
	@echo "opencode directly:  set -a; . ./.env; set +a"

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
	@test -f openspec/config.yaml || { echo "ERROR: OpenSpec not initialized. Run once and commit: openspec init --tools opencode --force"; exit 1; }
	docker pull ghcr.io/sooperset/mcp-atlassian:latest
	opencode run --agent spec-importer \
		"Import Confluence pages under parent ID $$CONFLUENCE_PARENT_ID into openspec/specs/ as OpenSpec spec files."

dev: ## Work a ticket headless in one shot with jira-dev (set JIRA_KEY)
	@test -n "$${JIRA_KEY}" || { echo "ERROR: set JIRA_KEY, e.g. make dev JIRA_KEY=KAN-1"; exit 1; }
	opencode run --agent jira-dev "Implement Jira ticket $${JIRA_KEY}"

tui: ## Launch the interactive opencode TUI with .env auto-loaded
	opencode

# ---------------------------------------------------------------------------
# App targets
# ---------------------------------------------------------------------------

