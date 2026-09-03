## MODIFIED Requirements

### Requirement: Message model
The system MUST represent a chat message with a unique identifier, an origin role (user, assistant, tool, or system), free-text content, an optional list of tool call requests, and optional response statistics consisting of an output-token count and elapsed generation duration for completed assistant answers.

#### Scenario: Legacy message data
- **WHEN** a persisted message has no response-statistics fields
- **THEN** the system MUST load it successfully without response statistics

#### Scenario: Completed answer data
- **WHEN** a completed final assistant answer has response statistics
- **THEN** the system MUST preserve those statistics when serializing and loading the session message list

### Requirement: Streaming last-message update
The system MUST support incrementally appending text to the last message of a session and broadcast a `last_message_update` event carrying the session ID and the updated message. When response statistics become available for a successful final assistant answer, the system MUST deliver them through that existing event.

#### Scenario: Final statistics update
- **WHEN** a successful final assistant answer receives response statistics after streaming text is complete
- **THEN** the system MUST broadcast its updated message and statistics through `last_message_update`
