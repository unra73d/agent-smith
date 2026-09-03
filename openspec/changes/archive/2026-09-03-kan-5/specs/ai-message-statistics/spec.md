## Purpose

Provides compact generation-speed and duration information for completed final assistant answers so users can assess response performance in the chat.

## ADDED Requirements

### Requirement: Capture final-answer response statistics
The system SHALL record output-token count and elapsed generation duration for each successful final assistant answer. It SHALL use provider-reported completion-token usage when available and SHALL otherwise estimate output tokens from the generated answer text.

#### Scenario: Provider usage is available
- **WHEN** a final assistant answer completes successfully and the provider reports completion-token usage
- **THEN** the system SHALL retain the reported completion-token count and elapsed generation duration on that answer

#### Scenario: Provider usage is unavailable
- **WHEN** a final assistant answer completes successfully without provider completion-token usage
- **THEN** the system SHALL retain a local output-token estimate and elapsed generation duration on that answer

### Requirement: Display final-answer response statistics
The UI SHALL render response statistics beneath a completed final assistant answer in compact muted text as `<whole-number>t/s <duration>`, where throughput is output tokens divided by elapsed seconds and duration uses compact seconds or minutes-and-seconds formatting. The statistics SHALL be visible only while the containing assistant message is hovered, matching that message's action-button interaction, while reserving their footer space so changing visibility does not shift surrounding messages.

#### Scenario: Completed answer has statistics
- **WHEN** a final assistant answer has a positive output-token count and positive elapsed duration
- **THEN** the UI SHALL reveal statistics beneath the answer only while its containing message is hovered, such as `30t/s 5m 33s`, without changing the layout of surrounding messages

#### Scenario: Statistics are unavailable
- **WHEN** an assistant message is still streaming, failed, cancelled, intermediate to a tool invocation, or has unusable statistics
- **THEN** the UI SHALL omit the statistics footer
