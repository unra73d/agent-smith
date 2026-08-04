## Purpose

Provides standard Makefile targets for the project's developer workflow — run, server, test, vet, and build — so the application can be started, verified, and compiled with consistent, documented commands.

## ADDED Requirements

### Requirement: Run targets

The project SHALL provide `make run` and `make server` targets that start the application from source.

#### Scenario: Run the desktop application
- **WHEN** a developer runs `make run` from the project root
- **THEN** the desktop application is started from source (`go run main.go`)

#### Scenario: Run the HTTP server only
- **WHEN** a developer runs `make server` from the project root
- **THEN** the application is started in server-only mode on port 8008 (`go run main.go --server --port 8008`)

### Requirement: Verification targets

The project SHALL provide `make test` and `make vet` targets that verify the Go source tree.

#### Scenario: Run the test suite
- **WHEN** a developer runs `make test` from the project root
- **THEN** the Go test suite runs against all packages and reports pass/fail

#### Scenario: Run the vet checks
- **WHEN** a developer runs `make vet` from the project root
- **THEN** Go vet analysis runs against all packages

### Requirement: Build target

The project SHALL provide a `make build` target that compiles the application binary.

#### Scenario: Build the binary
- **WHEN** a developer runs `make build` from the project root
- **THEN** a runnable binary named `agentsmith` is produced in the `build/` directory

### Requirement: Target discoverability

The project SHALL list the app targets in `make help` output so developers can discover available commands.

#### Scenario: Help shows app targets
- **WHEN** a developer runs `make help`
- **THEN** the output includes the `run`, `server`, `test`, `vet`, and `build` targets with short descriptions
