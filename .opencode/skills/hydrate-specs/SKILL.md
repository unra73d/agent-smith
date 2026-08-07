---
name: hydrate-specs
description: Import the current canonical specs from Confluence into openspec/specs/ using the Atlassian MCP, inside the CURRENT OpenCode session. Use before proposing a change so OpenSpec has a baseline to diff against. Never shells out to opencode or make.
license: MIT
metadata:
  author: agent-smith
  version: "1.0"
---

# Hydrate specs from Confluence (in-session)

Pull the current canonical specs from Confluence into `openspec/specs/` using the
Atlassian MCP tools available in this session. **Do this inline — never run
`make init_spec`, `opencode`, or `opencode run`. You are already inside an
OpenCode session; launching another one is not allowed.**

## Inputs
- The Confluence parent page id is in the `CONFLUENCE_PARENT_ID` environment
  variable (read it with `printenv CONFLUENCE_PARENT_ID`).

## Steps
1. Enumerate the parent page and its descendants with the Atlassian MCP
   (`confluence_get_page`, `confluence_get_page_children` / `confluence_search`).
2. For each descendant page that documents system behavior, fetch its content and
   write it to `openspec/specs/<capability-slug>/spec.md`, where
   `<capability-slug>` is a kebab-case name derived from the page title with a
   trailing " Spec" removed (e.g. page "Chat Sessions Spec" →
   `openspec/specs/chat-sessions/spec.md`).
3. Convert each page into the OpenSpec spec shape — a `## Purpose` section plus one
   or more `## Requirements`, each with a `### Requirement:` header and at least
   one `#### Scenario:`. Preserve the documented rules; do not invent requirements.
4. Skip pages that are not behavioral specs (e.g. `README`, meeting notes).
5. Report which pages you imported and which you skipped.

Do not modify application code, and do not touch `openspec/changes/`.
