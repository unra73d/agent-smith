.PHONY: help init_env init_spec dev tui check run server test vet build

# Auto-load .env (KEY=VALUE) and export to every recipe, if the file exists.
# This is why you don't need `set -a; . ./.env; set +a` before make targets.
-include .env

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
	@echo "claude directly:  set -a; . ./.env; set +a"

# ---------------------------------------------------------------------------
# AI SDLC targets  (Claude Code — see CLAUDE.md)
# ---------------------------------------------------------------------------

## Env the Atlassian MCP server needs. Fail fast if missing.
REQUIRED_ENV := ATLASSIAN_DOMAIN ATLASSIAN_EMAIL ATLASSIAN_API_TOKEN

check: ## Verify .env and tooling are ready for the AI SDLC targets
	@test -f .env || { echo "ERROR: no .env — run: make init_env"; exit 1; }
	@for v in $(REQUIRED_ENV); do \
		grep -qE "^$$v=.+" .env || { echo "ERROR: $$v is not set in .env"; exit 1; }; \
	done
	@command -v claude >/dev/null 2>&1 || { echo "ERROR: claude CLI not installed"; exit 1; }
	@command -v openspec >/dev/null 2>&1 || { echo "ERROR: openspec CLI not installed"; exit 1; }
	@command -v npx >/dev/null 2>&1 || { echo "ERROR: npx not installed (the Atlassian MCP server runs via npx)"; exit 1; }
	@test -f openspec/config.yaml || { echo "ERROR: OpenSpec not initialized. Run once and commit: openspec init --tools claude --force"; exit 1; }
	@echo "OK — ready. Try: make dev JIRA_KEY=KAN-1"

init_spec: check ## Hydrate openspec/specs from Confluence (source of truth)
	@grep -qE '^CONFLUENCE_PARENT_ID=.+' .env || { echo "ERROR: CONFLUENCE_PARENT_ID is not set in .env"; exit 1; }
	claude "Use the spec-hydrator subagent to hydrate openspec/specs/ from Confluence."

dev: check ## Work a ticket interactively with /implement (set JIRA_KEY)
	@test -n "$${JIRA_KEY}" || { echo "ERROR: set JIRA_KEY, e.g. make dev JIRA_KEY=KAN-1"; exit 1; }
	claude "/implement $${JIRA_KEY}"

tui: ## Launch an interactive Claude Code session in this project
	claude

# ---------------------------------------------------------------------------
# App targets
# ---------------------------------------------------------------------------

run: ## Run the desktop application
	go run main.go

server: ## Run only the HTTP server (no UI) on port 8008
	go run main.go --server --port 8008

test: ## Run the Go test suite
	go test ./...

vet: ## Run go vet analysis on all packages
	go vet ./...

build: ## Build the application binary into build/agentsmith
	go build -o build/agentsmith .

