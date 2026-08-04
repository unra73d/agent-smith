## 1. Makefile app targets

- [x] 1.1 Add `run` target (`go run main.go`) to the App targets section
- [x] 1.2 Add `server` target (`go run main.go --server --port 8008`)
- [x] 1.3 Add `test` target (`go test ./...`)
- [x] 1.4 Add `vet` target (`go vet ./...`)
- [x] 1.5 Add `build` target (`go build -o build/agentsmith .`)
- [x] 1.6 Update `.PHONY` and confirm all new targets appear in `make help`

## 2. Documentation

- [x] 2.1 Update README "Running" section to mention `make run` and `make server`

## 3. Verification

- [x] 3.1 Run `make vet`, `make test`, and `make build` and confirm all pass
- [x] 3.2 Run `make help` and confirm the new targets are listed
