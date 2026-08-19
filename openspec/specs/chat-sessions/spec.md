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
- THEN they are stored in a local SQLite database (app.db)
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
- THEN an SSE event (session_update, new_message, or last_message_update) is emitted to connected clients
