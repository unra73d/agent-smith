## Purpose

Define reusable roles that shape the assistant's behavior through composable system-prompt fields.

## Requirements

### Requirement: Users can manage roles

#### Scenario: CRUD on roles
- WHEN a user creates, lists, updates, or deletes a role
- THEN the operation succeeds and is reflected in the role list

### Requirement: A role composes the system prompt from fields

#### Scenario: System prompt assembled from role fields
- WHEN a role is applied to a conversation
- THEN its name plus General Instruction, Role, and Style fields together form the system prompt

### Requirement: Role changes are broadcast in real time

#### Scenario: SSE event on role change
- WHEN a role is added, updated, or removed
- THEN an SSE role_list_update event is emitted
