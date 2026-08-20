## Purpose

Run as a desktop application or a headless HTTP server, exposing a REST API and an SSE event stream.

## Requirements

### Requirement: The app runs as desktop or headless server

#### Scenario: Desktop and server modes

- By default it launches a desktop webview UI; with `--server` (and optional `--port`) it runs headless as an HTTP server.

### Requirement: A REST API exposes all agent features

#### Scenario: Agent feature endpoints

- Endpoints under `/agent` cover sessions, models, providers, roles, MCP servers, and the chat modes.

### Requirement: An SSE endpoint streams server events

#### Scenario: Server event stream

- The `/sse` endpoint streams session, message, provider, role, and MCP list updates plus periodic heartbeats.

### Requirement: External links open in the system browser

#### Scenario: Open an external URL

- A desktop helper endpoint opens a given URL in the user's default browser.
