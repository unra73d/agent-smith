---
name: knowledgebase
description: Read relevant Confluence knowledge-base context through the Atlassian MCP. Use when repository, Jira, or OpenSpec context points to architecture or behavioral documentation.
allowed-tools: mcp__atlassian__*
---

# Knowledgebase

Use the Atlassian MCP as a read-only Confluence context source. Never create,
edit, publish, move, or delete Confluence pages.

1. Locate pages referenced by the ticket, its comments, linked Jira items, current
specs, or clear code terminology. Read the sections relevant to the change.
2. Follow related Confluence links only when they are likely to clarify requirements,
architecture, integration contracts, or prior decisions. Avoid broad page-tree
enumeration without a concrete question.
3. Preserve the distinction between confirmed documented behavior and inference.
4. Return the consulted pages, relevant conclusions, conflicts with other sources,
and unresolved questions for the primary agent to discuss with the user.