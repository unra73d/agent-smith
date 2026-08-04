# Agent Smith

## Purpose

Agent Smith is an AI agent that can work with multiple LLM models, MCP servers, and AI providers that expose an OpenAI-compatible completions API. It is intended for PoC/MVP use cases where integrations need to be tested quickly or model performance evaluated. The agent runs as a desktop application by default and can also run as an HTTP server exposing a chat API over SSE.

## Requirements

### Requirement: Chat with LLM models

The agent shall provide a chat interface to an LLM model through configured AI providers that expose an OpenAI-compatible completions API. A user must be able to send messages to the model and receive responses within a chat session.

#### Scenario: User sends a message and receives a response

- **WHEN** a user sends a chat message to a configured model in a session
- **THEN** the model generates a response
- **AND** the response is delivered back to the user as part of the chat session

#### Scenario: Chat fails when no model is configured

- **WHEN** a user sends a chat message but no matching model is configured for the request
- **THEN** the chat request is not processed
- **AND** an error is reported

### Requirement: Change model, role, and tools mid-conversation

A user shall be able to change the model, the agent role, and the set of available tools during an active conversation without starting a new session, enabling dynamic interaction scenarios.

#### Scenario: Switch model mid-conversation

- **WHEN** a user changes the model used by an active conversation
- **THEN** subsequent messages in that conversation are processed by the newly selected model

#### Scenario: Switch role mid-conversation

- **WHEN** a user changes the role used by an active conversation
- **THEN** subsequent messages in that conversation are processed with the newly selected role's system prompt

#### Scenario: Switch tools mid-conversation

- **WHEN** a user changes the tool set available to an active conversation
- **THEN** subsequent messages in that conversation may invoke the newly available tools

### Requirement: Custom agent roles via system prompts

The agent shall support creating customized agent roles defined through system prompts, including a general instruction, a role description, and a text style and tone, so that the model behavior can be tailored per role.

#### Scenario: User creates a custom role

- **WHEN** a user defines a role with a general instruction, a role and personality description, and a text style and tone
- **THEN** the role is persisted
- **AND** the role becomes available for use in chat sessions

#### Scenario: Chat uses the role's system prompt

- **WHEN** a chat session uses a role
- **THEN** the model is prompted with the role's general instruction, role and personality description, and text style and tone

### Requirement: Use tools from MCP servers

The agent shall expose tools provided by MCP (Model Context Protocol) servers as callable tools, supporting both SSE and stdio transports, so that the model can invoke external functionality through configured MCP servers.

#### Scenario: Load tools from an MCP server

- **WHEN** an MCP server is configured and active
- **THEN** the agent connects to the server using its transport
- **AND** the server's tools are loaded and made available for invocation

#### Scenario: Connect via SSE transport

- **WHEN** an MCP server is configured with an SSE transport and URL
- **THEN** the agent connects to the server over SSE to discover and call its tools

#### Scenario: Connect via stdio transport

- **WHEN** an MCP server is configured with a stdio transport and command
- **THEN** the agent spawns the command and connects to the server over its standard input/output to discover and call its tools

#### Scenario: Tool invocation result is returned to the model

- **WHEN** the model invokes a tool provided by an MCP server
- **THEN** the tool result is returned to the model so it can reason about the outcome

### Requirement: Builtin Lua code execution tool

The agent shall provide a builtin tool that executes Lua code, so the model can perform precise calculations. Output produced with `print()` is accumulated and returned as the result, and the last value left on the stack is appended to the result.

#### Scenario: Model runs a Lua calculation

- **WHEN** the model invokes the builtin Lua code runner with Lua code
- **THEN** the code is executed
- **AND** the accumulated `print()` output, along with the last returned value, is returned as the tool result

### Requirement: Multiple chats in parallel

The agent shall support running multiple chats concurrently so that a user can maintain and interact with several conversations in parallel.

#### Scenario: User runs concurrent chats

- **WHEN** a user has multiple active chat sessions
- **THEN** the sessions are maintained independently
- **AND** each session can receive messages and produce responses without interfering with the others

### Requirement: Desktop and HTTP server run modes

By default the agent shall run as a desktop application. The agent shall also be runnable as an HTTP server only, exposing the agent part over HTTP, so it can be consumed programmatically.

#### Scenario: Run as desktop application

- **WHEN** the agent is started without server flags
- **THEN** it runs as a desktop application

#### Scenario: Run as HTTP server

- **WHEN** the agent is started with the `--server` flag and a `--port` argument (for example `--server --port 8008`)
- **THEN** only the agent part runs as an HTTP server on the specified port

### Requirement: Server API with SSE updates

When running as an HTTP server, the agent shall provide a Server API where clients connect to an SSE endpoint to receive updates from the server and use other APIs to invoke features.

#### Scenario: Client connects to the SSE endpoint

- **WHEN** a client connects to the SSE endpoint of the running server
- **THEN** the client receives a stream of updates emitted by the server

#### Scenario: Client invokes a feature via the API

- **WHEN** a client calls an API endpoint of the server
- **THEN** the requested feature is executed
- **AND** results are made available to the client, including updates delivered over the SSE connection
