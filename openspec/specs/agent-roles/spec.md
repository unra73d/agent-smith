## Purpose

Define reusable roles that shape the assistant's behavior through composable system-prompt fields.

## Requirements

### Requirement: Users can manage roles

#### Scenario: Roles can be managed

- Roles can be created, listed, updated, and deleted.

### Requirement: A role composes the system prompt from fields

#### Scenario: Role fields form the system prompt

- A role carries a name plus General Instruction, Role, and Style fields that together form the system prompt applied to a conversation.

### Requirement: Role changes are broadcast in real time

#### Scenario: Role changes emit an update

- Adding, updating, or removing a role emits an SSE `role_list_update` event.
