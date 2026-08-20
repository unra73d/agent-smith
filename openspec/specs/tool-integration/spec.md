## Purpose

Extend the assistant with external tools from MCP servers and a built-in Lua execution tool.

## Requirements

### Requirement: Users can manage MCP servers

#### Scenario: MCP server management and reload

- MCP servers can be created, listed, updated, deleted, tested, and reloaded (individually or all at once).
- Each has a name, transport, URL or command, and an active flag.

### Requirement: Both stdio and SSE MCP transports are supported

#### Scenario: MCP transport configuration

- An MCP server may be configured with a stdio command or an SSE URL.

### Requirement: Active MCP tools are offered to the LLM

#### Scenario: Active tools are available

- During tool chat, tools from active MCP servers are made available for the model to call.

### Requirement: A built-in Lua tool enables precise computation

#### Scenario: Deterministic Lua computation

- The model can execute Lua code through a built-in tool to compute results deterministically.

### Requirement: MCP changes are broadcast in real time

#### Scenario: MCP changes emit an update

- MCP server create/update/delete/reload emits an SSE `mcp_list_update` event.
