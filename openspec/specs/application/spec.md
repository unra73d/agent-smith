# Application Shell Specification

## Purpose

Agent Smith is a Go desktop/server application that orchestrates interactions between one or
more LLM models, OpenAI-compatible AI providers, MCP servers, and built-in tools. This
capability describes the process-level application shell: how the program boots, initializes
its storage and agent state, and how it can be run either as a native desktop application
(webview UI) or as a headless HTTP server.

## Requirements

### REQ-APP-001: Process entry point
The system MUST expose a single Go entry point (`main.go`) that is responsible for parsing
command-line flags, initializing the agent runtime, starting the HTTP server, and launching
the optional desktop UI.

### REQ-APP-002: Server-only flag
The system MUST support a `--server` boolean flag that, when set, runs only the HTTP server
without launching a native UI window, and blocks forever until the process receives a
termination signal.

### REQ-APP-003: Port override flag
The system MUST support an integer `--port` flag that overrides the port on which the HTTP
server listens; when unset (zero), the server MUST bind to an OS-assigned ephemeral port.

### REQ-APP-004: UI embedding
The system MUST embed the web UI directory (`src/ui`) into the compiled binary using Go
`embed` directives so the UI is served without an external filesystem dependency.

### REQ-APP-005: Database file location
The system MUST set the `AS_AGENT_DB_FILE` environment variable to `app.db` at startup and
MUST initialize the SQLite database schema if the database file does not already exist.

### REQ-APP-006: Agent state initialization
The system MUST load agent state on startup, including AI providers, historical chat
sessions, agent roles, MCP servers, and built-in tools, and MUST create a global in-memory
"flash" session used for transient operations.

### REQ-APP-007: Desktop UI
When not running in server-only mode, the system MUST open a native webview window titled
"Agent Smith", size it to 1000x800 pixels, and navigate it to the served UI URL
(`http://<addr>/ui/`).

### REQ-APP-008: Graceful shutdown
The system MUST handle `SIGINT`, `SIGTERM`, and OS interrupt signals, initiating a graceful
shutdown of the HTTP server with a bounded grace period before the process exits.