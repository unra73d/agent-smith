## Purpose

Extend the assistant with external tools from MCP servers and a built-in Lua execution tool.

## Requirements

### Requirement: Users can manage MCP servers

The system SHALL allow users to create, list, update, delete, test, and reload MCP servers, individually or all at once, each configured with a name, transport, URL or command, and an active flag.

#### Scenario: CRUD, test, and reload MCP servers
- WHEN a user creates, lists, updates, deletes, tests, or reloads an MCP server (individually or all at once)
- THEN the operation succeeds
- AND each MCP server has a name, transport, URL or command, and an active flag

### Requirement: Both stdio and SSE MCP transports are supported

The system SHALL support configuring an MCP server with either a stdio command or an SSE URL.

#### Scenario: Configure transport type
- WHEN an MCP server is configured
- THEN it may use a stdio command or an SSE URL

### Requirement: Active MCP tools are offered to the LLM

The system SHALL make tools from active MCP servers available for the model to call during tool chat.

#### Scenario: Tools available during tool chat
- WHEN a tool chat is running
- THEN tools from active MCP servers are made available for the model to call

### Requirement: A built-in Lua tool enables precise computation

The system SHALL provide a built-in Lua tool that the model can execute to compute results deterministically.

#### Scenario: Model executes Lua for deterministic results
- WHEN the model needs a precise computation
- THEN it can execute Lua code through a built-in tool to compute the result deterministically

### Requirement: MCP changes are broadcast in real time

The system SHALL broadcast MCP server changes to connected clients via an SSE mcp_list_update event.

#### Scenario: SSE event on MCP server change
- WHEN an MCP server is created, updated, deleted, or reloaded
- THEN an SSE mcp_list_update event is emitted
