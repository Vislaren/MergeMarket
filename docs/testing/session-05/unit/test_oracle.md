## Oracle — Mock Server (B-02) — Unit

The source of truth is `project_docs/api/API_CONTRACTS.md`, the B-02 task spec in
`Agent_B/TODO.md`, and the Go coding standards in `Agent_B/INSTRUCTIONS.md §3`.

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `go build ./...` | service is valid Go | exit 0 | Go toolchain |
| `go test ./...` | in-package tests | exit 0, all pass | INSTRUCTIONS §3 (testing) |
| `go vet` / `gofmt -l` | static quality | clean / empty | INSTRUCTIONS §3 |
| `go.mod` | offline-dev mock | no `require` block (stdlib-only) | B-02 design goal |
| `config.go` | env-driven config | reads `MOCK_SERVER_PORT`/`PORT`; `defaultPort = 8080` | INSTRUCTIONS §3; B-02 "port 8080" |
| `server.go` routes | full contract surface | all 13 `/api/v1/*` + `/health` registered | API_CONTRACTS.md |
| HealthResponse | health shape | `{status, service, version}` | API_CONTRACTS Health |
| Search offer | money math | `total_cost == price + shipping` | API_CONTRACTS Search |
| ErrorResponse | error shape | `{error, message}` | API_CONTRACTS Error Shape |
| `main.go` | logging | imports `log/slog` | INSTRUCTIONS §3 |
| Each package file | documentation | has GoDoc package/command comment | INSTRUCTIONS §3 |
