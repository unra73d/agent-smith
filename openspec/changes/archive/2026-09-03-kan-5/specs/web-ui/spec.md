## MODIFIED Requirements

### Requirement: Message rendering
The UI MUST render messages by origin, rendering assistant messages with collapsible thinking and tool sections, markdown content, and compact muted response statistics beneath a completed final answer when supplied. Statistics MUST be visible only while the containing assistant message is hovered, matching message action buttons, while retaining their layout space to prevent surrounding messages from moving; tool messages render as collapsible tool responses; and user messages render as plain text.

#### Scenario: Render completed answer statistics
- **WHEN** an assistant message contains usable response statistics
- **THEN** the UI MUST render the statistics beneath its answer without including them in rendered markdown content, reveal them only while the containing message is hovered, and preserve surrounding-message layout while visibility changes

#### Scenario: Render an answer without statistics
- **WHEN** an assistant message has no usable response statistics
- **THEN** the UI MUST render the message without a statistics footer

### Requirement: Real-time message updates
The UI MUST react to `new_message`, `last_message_update`, and `session_update` SSE events by appending messages, updating the in-progress assistant message, replacing updated sessions, and retaining and rendering response statistics carried by assistant messages.

#### Scenario: Final response-statistics update
- **WHEN** a `last_message_update` event carries response statistics for the active assistant message
- **THEN** the UI MUST update that message's statistics footer without disturbing its rendered markdown content
