## Purpose

Run as a desktop application or a headless HTTP server, exposing a REST API and an SSE event stream.

## Requirements

### Requirement: The app runs as desktop or headless server

#### Scenario: Default desktop launch
- WHEN the app starts with no flags
- THEN it launches a desktop webview UI

#### Scenario: Headless server launch
- WHEN the app starts with --server (and optional --port)
- THEN it runs headless as an HTTP server

### Requirement: A REST API exposes all agent features

#### Scenario: Endpoints under /agent
- WHEN a client calls the REST API
- THEN endpoints under /agent cover sessions, models, providers, roles, MCP servers, and the chat modes

### Requirement: An SSE endpoint streams server events

#### Scenario: /sse streams updates and heartbeats
- WHEN a client connects to /sse
- THEN it receives session, message, provider, role, and MCP list updates plus periodic heartbeats

### Requirement: External links open in the system browser

#### Scenario: Desktop helper opens URL
- WHEN a given URL is passed to the desktop helper endpoint
- THEN it opens in the user's default browser
