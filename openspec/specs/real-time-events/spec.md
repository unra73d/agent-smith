# Real-Time Events Specification

## Purpose

Agent Smith keeps connected clients in sync with server state through a single Server-Sent
Events (SSE) stream. This capability describes the SSE connection model, the typed event
broadcast bus, and the set of events the system emits so the UI can update reactively.

## Requirements

### REQ-RTE-001: SSE endpoint
The system MUST expose a `GET /agent/sse` endpoint that returns a long-lived Server-Sent
Events stream over which updates are pushed to connected clients.

### REQ-RTE-002: Event broadcast bus
The system MUST maintain a single in-process broadcast channel through which components emit
typed update messages consumed by the SSE handler.

### REQ-RTE-003: Event types
The system MUST support emitting the following SSE event types:
`session_update`, `new_message`, `last_message_update`, `provider_list_update`,
`mcp_list_update`, and `role_list_update`.

### REQ-RTE-004: Session update event
The system MUST emit a `session_update` event carrying the updated session object whenever a
session's messages, title, or other state changes.

### REQ-RTE-005: New message event
The system MUST emit a `new_message` event carrying the newly appended message and its
session ID whenever a message is added to a session.

### REQ-RTE-006: Last message update event
The system MUST emit a `last_message_update` event carrying the session ID and the updated
last message whenever streaming text is appended to the current assistant message.

### REQ-RTE-007: List update events
The system MUST emit `provider_list_update`, `mcp_list_update`, and `role_list_update` events
carrying the full corresponding collection whenever providers, MCP servers, or roles are
created, updated, or deleted.

### REQ-RTE-008: Event serialization
The system MUST serialize each event's payload to JSON before dispatch and MUST skip emitting
an event whose payload cannot be serialized.

### REQ-RTE-009: Heartbeat
The system MUST emit a periodic `heartbeat` event (roughly every 10 seconds) to keep the SSE
connection alive and detect dead connections.

### REQ-RTE-010: Connection lifecycle
The system MUST close the SSE stream when the underlying channel is closed or the client
disconnects.