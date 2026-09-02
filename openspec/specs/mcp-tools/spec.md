# MCP Servers & Tools Specification

## Purpose

Agent Smith integrates external tools exposed by MCP (Model Context Protocol) servers. This
capability covers the MCP server data model, supported transports, the CRUD lifecycle,
connectivity testing and reload, asynchronous tool discovery, tool calling, and how tools are
selected and deduplicated for the model.

## Requirements

### REQ-MCP-001: MCP server model
The system MUST represent an MCP server with a unique ID, name, transport type, URL, command,
a flag indicating whether it is active, a flag indicating whether its tools have been loaded,
and a list of discovered tools.

### REQ-MCP-002: Supported transports
The system MUST support three MCP transports: `stdio` (with a shell command), `sse`, and
`http` (streamable HTTP), and MUST construct the appropriate MCP client for the configured
transport.

### REQ-MCP-003: Server loading
The system MUST load all persisted MCP servers from the `mcp` table on startup and asynchronously
load each server's tools.

### REQ-MCP-004: Server persistence
The system MUST persist MCP servers using an upsert into an `mcp` table keyed by ID, storing
id, name, transport, url, command, and the active flag.

### REQ-MCP-005: Server creation
The system MUST create an MCP server with a fresh UUID, persist it, add it to the in-memory
list, load its tools asynchronously, and broadcast a `mcp_list_update` event.

### REQ-MCP-006: Server update
The system MUST update an MCP server's name and active flag, and when its transport, URL, or
command changed, MUST clear its loaded tools, reload them asynchronously, persist, and
broadcast `mcp_list_update` events.

### REQ-MCP-007: Server deletion
The system MUST delete an MCP server by ID from storage and the in-memory list and broadcast
a `mcp_list_update` event.

### REQ-MCP-008: Server reload
The system MUST support reloading a single server's tools (clearing existing tools, marking it
unloaded, reloading asynchronously) and reloading all servers' tools, broadcasting
`mcp_list_update` events.

### REQ-MCP-009: Connectivity test
The system MUST test an MCP server by loading its tools, considering the server valid only
when loading succeeds and at least one tool is discovered.

### REQ-MCP-010: Tool discovery
The system MUST discover a server's tools by issuing an MCP `tools/list` request and mapping
each tool's name, description, input schema properties (type and description), and required
parameters into the internal tool model.

### REQ-MCP-011: Tool call
The system MUST call a tool by issuing an MCP `tools/call` request with the tool name and
parameters, returning the concatenated text content of the result, and MUST treat a tool that
flags `isError` by passing its content through rather than aborting.

### REQ-MCP-012: Tool call error handling
When a tool call fails, the system MUST extract any available text content, or when the
server returned a bare error with missing content, return a descriptive placeholder as the
tool result so the model can reason about it.

### REQ-MCP-013: Tool call timeout
The system MUST bound tool discovery and tool calls with timeouts (tool load and MCP init
respectively), returning an error when the deadline is exceeded.

### REQ-MCP-014: Tool exposure
The system MUST expose the union of tools from all active MCP servers plus the built-in
tools, and MUST deduplicate by tool name so that the first active server owning a name wins
and each tool resolves to the server that listed it.

### REQ-MCP-015: Tool-to-server resolution
The system MUST resolve a requested tool name to the MCP server that owns it so tool calls are
routed to the correct server.