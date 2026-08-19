## Purpose

Provide multiple ways to converse with an LLM: a direct streaming chat, a tool-enabled agent chat, and a dynamic self-as-tool agent.

## Requirements

### Requirement: Direct chat streams a plain model response

The system SHALL stream the model's response over SSE in direct mode without invoking any tools.

#### Scenario: Streamed response with no tools
- WHEN a message is sent in direct mode
- THEN the model response is streamed over SSE
- AND no tools are invoked
- AND the call completes when generation ends

### Requirement: Tool chat runs an agent loop that can call tools

The system SHALL run an agent loop in tool mode that allows the LLM to call tools or dynamic agents until the LLM finishes, the user intervenes, or recursion is detected.

#### Scenario: Agent loop calls tools or dynamic agents
- WHEN a message is sent in tool mode
- THEN the LLM may call tools or dynamic agents
- AND the loop ends when the LLM finishes, the user intervenes, or recursion is detected

### Requirement: Dynamic agent chat is callable as a tool

The system SHALL expose a dynamic agent, configured only by a system prompt, as a callable tool that returns a complete non-streamed message, may call tools recursively while guarding recursion depth, and persists no messages or sessions.

#### Scenario: Dynamic agent behavior
- WHEN a dynamic agent is invoked
- THEN it is configured only by a system prompt
- AND it returns a complete, non-streamed message
- AND it may call tools recursively while guarding recursion depth
- AND it persists no messages or sessions

### Requirement: Users can reconfigure a conversation mid-flight

The system SHALL allow users to change the model, role, and available tools partway through a conversation, taking effect on the next message.

#### Scenario: Change model, role, or tools mid-conversation
- WHEN a user changes the model, role, or available tools partway through a conversation
- THEN the change takes effect on the next message
