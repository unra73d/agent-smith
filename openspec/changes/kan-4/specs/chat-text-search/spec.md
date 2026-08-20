## Purpose

Provide a fast, chat-scoped find experience that lets users locate and navigate literal text across the active conversation without contacting the backend.

## ADDED Requirements

### Requirement: Users can search the active chat locally

The system SHALL provide a find panel for the active chat that accepts a literal, case-sensitive search string and searches only visible text in user messages and final model answers. Tool calls, tool responses, and thinking/reasoning content SHALL be excluded whether their collapsible nodes are expanded or collapsed.

#### Scenario: Open search from the keyboard
- **WHEN** the user presses Ctrl+F or Cmd+F while the chat is active
- **THEN** the chat find panel opens in the top-right and focuses its search input
- **AND** the browser or webview's default page-search action does not open

#### Scenario: Search after typing delay
- **WHEN** the user changes the search string
- **THEN** matching is performed after 500 milliseconds without further input
- **AND** no backend or chat API request is made

#### Scenario: Literal case-sensitive matching
- **WHEN** the search string contains ordinary characters, whitespace, or regex metacharacters
- **THEN** the system treats the string literally and matches only identical casing

#### Scenario: Search excludes non-conversational nodes
- **WHEN** the search string occurs only in a tool call, tool response, or thinking/reasoning node
- **THEN** the system reports zero matches

#### Scenario: Search includes visible user and final-answer text
- **WHEN** the search string occurs in a visible user message or final model answer
- **THEN** the system includes that occurrence in the match count and highlights it

#### Scenario: Empty or absent search text
- **WHEN** the search string is empty
- **THEN** all search highlights are removed
- **AND** the panel reports zero matches and no active match

### Requirement: Search results are highlighted and navigable

The system SHALL highlight every match with a muted color, distinguish the active match with a separate color, display the total match count, and provide previous and next controls.

#### Scenario: Matches are displayed
- **WHEN** the settled search string matches text in the active chat
- **THEN** every occurrence is highlighted with the muted match style
- **AND** the panel displays the total number of occurrences
- **AND** the first occurrence becomes the active match

#### Scenario: Navigate to the next or previous match
- **WHEN** the user activates the next or previous control
- **THEN** the active match advances or rewinds with wraparound
- **AND** the active match uses the active style
- **AND** the active match is scrolled into view

#### Scenario: No matches
- **WHEN** the settled search string does not occur in the active chat
- **THEN** the panel displays zero matches
- **AND** navigation controls are unavailable

### Requirement: Search can be dismissed and is cancelled by new messages

The system SHALL allow the user to dismiss the find panel and SHALL keep search state consistent with active-session and streamed-message changes.

#### Scenario: Close search
- **WHEN** the user activates the panel close control or presses Escape while the panel is open
- **THEN** the panel closes
- **AND** all search highlights are removed

#### Scenario: Submit a new message
- **WHEN** the user submits a new message while the find panel is open or a search debounce is pending
- **THEN** the search is cancelled
- **AND** the pending search is prevented from running
- **AND** the find panel closes
- **AND** all search highlights are removed

#### Scenario: Active session changes
- **WHEN** the active chat session changes while search is open
- **THEN** the current query is applied to the newly active chat
- **AND** the match count and active match are recalculated

#### Scenario: Chat content changes
- **WHEN** an assistant message is updated while a query remains active
- **THEN** the query is reapplied to the current eligible rendered chat content
- **AND** the active match remains valid or resets to the first current match
