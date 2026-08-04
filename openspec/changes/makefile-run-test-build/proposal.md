## Why

The project's Makefile has an empty "App targets" section: `run`, `server`, `test`,
and `build` were advertised (`.PHONY` lists `run`/`server` and
`docs/ai-sdlc-setup.md` references `run`/`server`) but never implemented. Without
these targets there is no consistent, documented way to run, test, or build the
Go application, and CI tooling cannot reuse a standard set of commands.

## What Changes

- Add a `run` target that starts the desktop application (`go run main.go`).
- Add a `server` target that starts the HTTP server only (`go run main.go --server --port 8008`).
- Add a `test` target that runs the Go test suite (`go test ./...`).
- Add a `vet` target that runs `go vet ./...`.
- Add a `build` target that compiles the binary to `build/agentsmith` (already gitignored).
- Document the new targets in `help` output and README running instructions.

## Capabilities

### New Capabilities
- `project-tooling`: Standard Makefile targets for the developer workflow — run, server, test, vet, build — so the project can be run, verified, and compiled with consistent commands.

### Modified Capabilities
<!-- none: openspec/specs/ is empty, no existing capability specs to modify -->

## Impact

- `Makefile`: new targets in the "App targets" section, `.PHONY` updated.
- `README.md`: "Running" section updated to mention `make run` / `make server`.
- No application code changes; Go baseline (`go build`, `go vet`, `go test`) is already green.
- No dependency changes.
