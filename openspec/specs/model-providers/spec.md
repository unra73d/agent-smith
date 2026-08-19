## Purpose

Connect to any OpenAI-compatible completions API as a configurable provider and expose its models to conversations.

## Requirements

### Requirement: Users can manage providers

#### Scenario: CRUD and connectivity test on providers
- WHEN a user creates, lists, updates, deletes, or connectivity-tests a provider
- THEN the operation succeeds
- AND each provider has a name, API URL, API key, and rate limit

### Requirement: Available models are listed across providers

#### Scenario: Union of models across configured providers
- WHEN selecting a model for a conversation
- THEN the union of models offered by all configured providers is available for selection

### Requirement: Per-provider rate limits are enforced

#### Scenario: Requests respect configured rate limit
- WHEN requests are sent to a provider
- THEN they respect that provider's configured rate limit

### Requirement: Provider changes are broadcast in real time

#### Scenario: SSE event on provider change
- WHEN a provider is created, updated, or deleted
- THEN an SSE provider_list_update event is emitted
