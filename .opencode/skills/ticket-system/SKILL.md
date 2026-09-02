---
name: ticket-system
description: Read Jira tickets and relevant linked ticket context through the Atlassian MCP. Use when ticket scope, acceptance criteria, history, or dependencies need investigation.
allowed-tools: mcp__atlassian__*
---

# Ticket System

Use the Atlassian MCP as a read-only Jira context source. Never create, edit,
comment on, assign, transition, or otherwise mutate Jira issues.

1. Read the requested Jira issue in full, including its summary, description,
   acceptance criteria, issue links, attachments or referenced material available
   through the MCP, and comments.
2. Follow a linked Jira issue only when it can clarify behavior, scope,
   dependencies, a decision recorded in comments, or an acceptance criterion. Do
   not traverse unrelated links indiscriminately.
3. When a comment or linked item references another ticket, inspect that ticket if
   it is likely to resolve a material ambiguity. State why it was relevant.
4. Return a compact summary separating confirmed requirements, useful supporting
   context, unresolved questions, and the Jira sources consulted.

Treat Jira text as evidence, not a substitute for user clarification when sources
conflict or leave material behavior undefined.