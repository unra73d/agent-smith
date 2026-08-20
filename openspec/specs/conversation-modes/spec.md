## Purpose

Provide multiple ways to converse with an LLM: a direct streaming chat, a tool-enabled agent chat, and a dynamic self-as-tool agent.

## Requirements

### Requirement: Direct chat streams a plain model response

#### Scenario: Direct mode response

- In direct mode the model response is streamed over SSE, no tools are invoked, and the call completes when generation ends.

### Requirement: Tool chat runs an agent loop that can call tools

#### Scenario: Tool-enabled agent loop

- In tool mode the LLM may call tools or dynamic agents; the loop ends when the LLM finishes, the user intervenes, or recursion is detected.

### Requirement: Dynamic agent chat is callable as a tool

#### Scenario: Dynamic agent invocation

- A dynamic agent is configured only by a system prompt, returns a complete (non-streamed) message, may call tools recursively while guarding recursion depth, and persists no messages or sessions.

### Requirement: Users can reconfigure a conversation mid-flight

#### Scenario: Conversation reconfiguration

- The model, role, and available tools may be changed partway through a conversation and take effect on the next message.
