# Built-in Tools Specification

## Purpose

In addition to MCP-served tools, Agent Smith ships built-in tools available to the model
without any external server. This capability describes the built-in tool set and the Lua code
execution tool that lets the model perform calculations.

## Requirements

### REQ-BIT-001: Built-in tool registration
The system MUST provide a set of built-in tools registered in code and returned to the agent
as part of the model's available tool set, in addition to MCP tools.

### REQ-BIT-002: Lua code runner tool
The system MUST expose a built-in tool named `lua_code_runner` described as executing Lua 5.1
code and returning results, with a single required string parameter `code`.

### REQ-BIT-003: Lua execution
The system MUST execute the supplied Lua 5.1 code in a fresh Lua state and capture the output
of the overridden `print()` function, plus the last value left on the stack as a JSON-encoded
`result`.

### REQ-BIT-004: Lua result composition
The system MUST return captured print output as the tool result, appending the JSON-encoded
last returned value when present, and MUST return the JSON result alone when no print output
was captured.

### REQ-BIT-005: Lua value serialization
The system MUST translate Lua values into JSON-friendly Go values, mapping numbers, strings,
booleans, sequential tables to arrays, and map-like tables to objects, and MUST represent
unsupported types (functions, userdata, channels, states) as descriptive strings.

### REQ-BIT-006: Lua state lifecycle
The system MUST create a fresh Lua state for each execution and ensure the state is closed
even when the code errors, and MUST append an error message when execution fails.