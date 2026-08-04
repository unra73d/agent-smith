## Purpose

Lets users find a specific conversation in the chat history by typing into a
search field at the top of the chats list, filtering sessions by title or
message content in real time.

## ADDED Requirements

### Requirement: Search field at the top of the chats list

The chat history SHALL provide a single-line search field at the top of the
chats list. The field SHALL accept free-text input and SHALL NOT require a
submit action.

#### Scenario: Search field is visible above the chats

- **WHEN** the user opens the Chat sessions tab
- **THEN** a single-line search field is displayed at the top of the chats list

### Requirement: Real-time chat filtering

The chats list SHALL filter in real time as the user types in the search field.
A chat SHALL be included in the filtered results when the query matches the
chat title (session summary) or the text of any message inside the chat. The
match SHALL be case-insensitive.

#### Scenario: Filtering by chat title

- **WHEN** the user types text into the search field that matches the title of a chat
- **THEN** only chats whose title contains the typed text are shown in the list

#### Scenario: Filtering by message text inside a chat

- **WHEN** the user types text into the search field that matches the content of a message inside a chat
- **THEN** that chat is shown in the filtered list even when its title does not contain the typed text

#### Scenario: No matches

- **WHEN** the user types text that matches no chat title and no message text
- **THEN** the chats list is empty

### Requirement: Clear search with inline button

The search field SHALL provide a clear button inside the field that resets the
search text and restores the full chats list with a single click.

#### Scenario: Clearing the search restores all chats

- **WHEN** the user clicks the clear button inside the search field
- **THEN** the search field text is cleared
- **AND** the full, unfiltered chats list is restored
