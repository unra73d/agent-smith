# Logging Specification

## Purpose

Agent Smith uses a small, leveled logging facility with per-component tags. This capability
describes the log levels, global and per-logger enable flags, tagged instances, and the error
checking helpers used throughout the application.

## Requirements

### REQ-LOG-001: Log levels
The system MUST support three log levels: DEBUG, WARN, and ERROR, each with a global enable
flag that gates printing.

### REQ-LOG-002: Tagged loggers
The system MUST support creating named logger instances (e.g., `agent`, `ai`, `tools`,
`server`) that prepend a tag to each message and carry per-instance enable flags for each
level, which MUST be combined with the global flags before printing.

### REQ-LOG-003: Leveled log methods
Each logger MUST provide `D`, `W`, and `E` methods that print only when both the instance flag
and the corresponding global flag for that level are enabled.

### REQ-LOG-004: Global log methods
The system MUST provide package-level `D`, `W`, and `E` functions that print only when the
corresponding global enable flag is set.

### REQ-LOG-005: Warning error check
The system MUST provide a `CheckW` helper that logs a warning with context when a supplied
error is non-nil and MUST always return whether the error was non-nil regardless of log level.

### REQ-LOG-006: Fatal error check
The system MUST provide a `CheckE` helper that, when a supplied error is non-nil, logs the
error (when error logging is enabled) and panics with the error, regardless of log level.

### REQ-LOG-007: Multi-error check
The system MUST provide a `CheckMultiE` helper that finds the first non-nil error from a
slice, logs each non-nil error, and panics with the first error.

### REQ-LOG-008: Panic recovery
The system MUST provide a `BreakOnError` recovery helper intended to be used with `defer`,
which catches panics raised by the fatal check helpers and logs them instead of terminating
the program.