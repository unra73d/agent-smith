# Persistence Specification

## Purpose

Agent Smith persists its state in a SQLite database located at the path named by the
`AS_AGENT_DB_FILE` environment variable. This capability describes the database schema, the
storage tables, and the persistence behavior of sessions, AI providers, roles, and MCP
servers.

## Requirements

### REQ-PER-001: Database location
The system MUST store all persistent state in a single SQLite database file whose path is
read from the `AS_AGENT_DB_FILE` environment variable, and MUST use the Go SQLite3 driver.

### REQ-PER-002: Schema initialization
The system MUST create the database schema idempotently (using `CREATE TABLE IF NOT EXISTS`)
covering all tables, and MUST be idempotent so it may be re-run without corrupting existing
data.

### REQ-PER-003: Sessions table
The system MUST maintain a `sessions` table with columns `session_id` (primary key), `date`,
`summary`, and `data` (JSON-encoded messages).

### REQ-PER-004: Providers table
The system MUST maintain a `providers` table with columns `id` (primary key), `name`,
`api_url`, `api_key`, `provider` (type), and `rate_limit`.

### REQ-PER-005: Roles table
The system MUST maintain a `roles` table with columns `id` (primary key) and `data`
(JSON-encoded role configuration).

### REQ-PER-006: MCP table
The system MUST maintain an `mcp` table with columns `id` (primary key), `name`, `transport`,
`url`, `command`, and `active` (boolean, default false).

### REQ-PER-007: Upsert semantics
The system MUST persist entities using `INSERT ... ON CONFLICT DO UPDATE` semantics so that
creating or updating an existing key overwrites the prior row rather than duplicating it.

### REQ-PER-008: Deletion
The system MUST delete rows by primary key for sessions, providers, roles, and MCP servers
when those entities are removed.

### REQ-PER-009: Corrupt-row tolerance
When loading entities, the system MUST tolerate rows that fail to parse or scan by skipping
them and logging a warning, continuing to load the remaining valid rows.