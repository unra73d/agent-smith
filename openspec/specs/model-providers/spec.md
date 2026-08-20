## Purpose

Connect to any OpenAI-compatible completions API as a configurable provider and expose its models to conversations.

## Requirements

### Requirement: Users can manage providers

#### Scenario: Provider management and connectivity

- Providers can be created, listed, updated, deleted, and connectivity-tested.
- Each provider has a name, API URL, API key, and rate limit.

### Requirement: Available models are listed across providers

#### Scenario: Models are available for selection

- The union of models offered by configured providers is available for selection in a conversation.

### Requirement: Per-provider rate limits are enforced

#### Scenario: Provider requests respect limits

- Requests to a provider respect its configured rate limit.

### Requirement: Provider changes are broadcast in real time

#### Scenario: Provider changes emit an update

- Provider create/update/delete emits an SSE `provider_list_update` event.
