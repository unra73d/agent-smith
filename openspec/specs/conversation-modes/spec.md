## Purpose

Provide multiple ways to converse with an LLM: a direct streaming chat, a tool-enabled agent chat, and a dynamic self-as-tool agent.

## Requirements

### Requirement: Direct chat streams a plain model response

#### Scenario: Streamed response with no tools
- WHEN a message is sent in direct mode
- THEN the model response is streamed over SSE
- AND no tools are invoked
- AND the call completes when generation ends

### Requirement: Tool chat runs an agent loop that can call tools

#### Scenario: Agent loop calls tools or dynamic agents
- WHEN a message is sent in tool mode
- THEN the LLM may call tools or dynamic agents
- AND the loop ends when the LLM finishes, the user intervenes, or recursion is detected

### Requirement: Dynamic agent chat is callable as a tool

#### Scenario: Dynamic agent behavior
- WHEN a dynamic agent is invoked
- THEN it is configured only by a system prompt
- AND it returns a complete, non-streamed message
- AND it may call tools recursively while guarding recursion depth
- AND it persists no messages or sessions

### Requirement: Users can reconfigure a conversation mid-flight

#### Scenario: Change model, role, or tools mid-conversation
- WHEN a user changes the model, role, or available tools partway through a conversation
- THEN the change takes effect on the next message
