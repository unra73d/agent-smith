## Purpose

Connect to any OpenAI-compatible completions API as a configurable provider and expose its models to conversations.

## Requirements

### Requirement: Users can manage providers

The system SHALL allow users to create, list, update, delete, and connectivity-test providers, each configured with a name, API URL, API key, and rate limit.

#### Scenario: CRUD and connectivity test on providers
- WHEN a user creates, lists, updates, deletes, or connectivity-tests a provider
- THEN the operation succeeds
- AND each provider has a name, API URL, API key, and rate limit

### Requirement: Available models are listed across providers

The system SHALL make the union of models offered by all configured providers available for selection in a conversation.

#### Scenario: Union of models across configured providers
- WHEN selecting a model for a conversation
- THEN the union of models offered by all configured providers is available for selection

### Requirement: Per-provider rate limits are enforced

The system SHALL enforce each provider's configured rate limit on requests sent to that provider.

#### Scenario: Requests respect configured rate limit
- WHEN requests are sent to a provider
- THEN they respect that provider's configured rate limit

### Requirement: Provider changes are broadcast in real time

The system SHALL broadcast provider changes to connected clients via an SSE provider_list_update event.

#### Scenario: SSE event on provider change
- WHEN a provider is created, updated, or deleted
- THEN an SSE provider_list_update event is emitted
