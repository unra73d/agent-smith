# HTTP Server & API Specification

## Purpose

Agent Smith exposes an HTTP API over a Gin router. The server serves the embedded web UI,
hosts REST endpoints under the `/agent` prefix, streams real-time updates over a Server-Sent
Events (SSE) endpoint, and provides CORS handling so the UI and external clients can consume
the API. In debug builds it additionally exposes diagnostic routes.

## Requirements

### REQ-HTTP-001: Gin server lifecycle
The system MUST run an HTTP server backed by the Gin framework that binds to the loopback
address (`127.0.0.1`) on either the requested port or an ephemeral port, and MUST report the
bound address back to the caller once the server is ready to accept requests.

### REQ-HTTP-002: Gin modes
The system MUST select Gin's debug mode when the global `DEBUG` logging flag is enabled and
release mode otherwise, and MUST register a panic-recovery middleware and the global CORS
middleware on all requests.

### REQ-HTTP-003: Static UI serving
The system MUST serve the embedded web UI assets under the `/ui/` path from the embedded
filesystem using Go's `http.FS`.

### REQ-HTTP-004: CORS
The system MUST set permissive CORS headers on responses: `Access-Control-Allow-Origin: *`,
allow the methods `GET,POST,PUT,PATCH,DELETE,OPTIONS`, and the headers `authorization,
origin, content-type, accept`. Preflight `OPTIONS` requests MUST be answered with HTTP 200
and not passed to downstream handlers.

### REQ-HTTP-005: Agent route prefix
The system MUST register all agent API routes under the `/agent` URL group prefix.

### REQ-HTTP-006: Request binding validation
For handlers that accept a JSON body or URI parameters, the system MUST bind the request to a
typed struct with Gin binding tags and MUST respond with HTTP 400 when required fields are
missing or binding fails.

### REQ-HTTP-007: Streaming request completion
For streaming endpoints, the system MUST return an HTTP 200 response when the underlying
stream completes successfully and HTTP 500 when the stream times out or errors.

### REQ-HTTP-008: Static file transport
The system MUST serve the built binary's web UI from the embedded filesystem so no separate
static file server is required.

### REQ-HTTP-009: Debug diagnostic routes
When debug mode is enabled, the system MUST register a `GET /debug/quit` route that gracefully
shuts down the server and a `GET /debug/initdb` route that (re)initializes the SQLite schema.

### REQ-HTTP-010: Client cancellation
The system MUST honor HTTP client disconnection by cancelling the request context so
in-flight generation and streaming work is aborted when the client disconnects.