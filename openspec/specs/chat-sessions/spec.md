## Purpose

Manage multiple independent chat conversations (sessions) that persist across app restarts and run in parallel.

## Requirements

### Requirement: Users can create, list, and delete sessions

The system SHALL allow users to create, list, and delete chat sessions.

#### Scenario: Create and delete a session

- WHEN a user creates a session
- THEN it appears in the session list
- WHEN the session is deleted
- THEN the session and its messages are removed

### Requirement: Multiple sessions run in parallel

The system SHALL support multiple sessions running in parallel, each with its own message history.

#### Scenario: Independent concurrent sessions

- WHEN several sessions are active at once
- THEN each has its own message history without interfering with one another

### Requirement: Sessions and messages persist locally

The system SHALL persist sessions and messages locally so history survives an application restart.

#### Scenario: History survives a restart

- WHEN conversations occur
- THEN they are stored in a local SQLite database (`app.db`)
- AND history survives an application restart

### Requirement: Users can prune history

The system SHALL allow users to truncate a session from a chosen message onward and to delete individual messages.

#### Scenario: Truncate from a message onward

- WHEN a user chooses a message in a session
- THEN the session can be truncated from that message onward

#### Scenario: Delete an individual message

- WHEN a user deletes an individual message
- THEN only that message is removed from the session

### Requirement: Session activity is broadcast in real time

The system SHALL broadcast session and message changes to connected clients via SSE events.

#### Scenario: SSE events on session and message changes

- WHEN a session or message changes
- THEN an SSE event (`session_update`, `new_message`, or `last_message_update`) is emitted to connected clients

### Requirement: Session titles are auto-summarized from conversation content

The system SHALL generate a short, human-readable title (2-4 words) summarizing the session's conversation, replacing any placeholder or raw-message-derived summary. The title SHALL be (re)generated after each of the first 3 completed user/assistant exchanges in a session, then SHALL remain stable and not be regenerated for any subsequent exchange.

An exchange is considered complete when the assistant has produced its final answer for that turn, with no further pending tool calls.

#### Scenario: Title generated after the first exchange

- WHEN a user sends their first message and the assistant's reply completes
- THEN the system generates a 2-4 word title summarizing the conversation
- AND persists it as the session's summary
- AND broadcasts the updated summary via the existing `session_update` SSE event

#### Scenario: Title regenerated through the second and third exchanges

- WHEN the assistant's reply completes for the session's 2nd or 3rd user message
- THEN the system regenerates the title from the conversation so far
- AND persists and broadcasts the updated summary

#### Scenario: Title stabilizes after the third exchange

- WHEN the assistant's reply completes for the session's 4th or later user message
- THEN the system does not regenerate the title
- AND the existing summary is left unchanged

#### Scenario: Tool and empty messages are excluded from summarization input

- WHEN the system builds the conversation content used to generate or regenerate a title
- THEN tool-call and tool-result messages are excluded
- AND messages with empty or whitespace-only text are excluded

#### Scenario: Reasoning/thinking content is excluded from the generated title

- WHEN the model's title-generation response includes reasoning or thinking content preceding the final answer
- THEN the persisted and broadcast summary contains only the model's final output, with the reasoning content excluded

#### Scenario: Title generation does not delay the chat response

- WHEN an exchange completes and the session is eligible for title (re)generation
- THEN the chat response already sent to the user is not delayed by the title generation call

#### Scenario: Temporary sessions are not titled

- WHEN a session is temporary (not persisted to the sessions list)
- THEN the system does not generate or broadcast a title for it
