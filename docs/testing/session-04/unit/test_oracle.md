## Oracle — A-04 Proxy-Validator Service (unit / structural)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `go build ./...` | service dir | exit 0 | Go compiles / A-04 deliverable |
| `go test ./...` | service dir | all packages `ok` | Agent A developer suite |
| `go vet ./...` | service dir | no diagnostics | Go standards |
| `gofmt -l .` | service dir | empty (no unformatted files) | INSTRUCTIONS §3 / SonarQube |
| `config.go` | — | reads env via `os.LookupEnv`; no hardcoded config | INSTRUCTIONS §3 ("config from env, never hardcoded") |
| `main.go` listen addr | — | `":"+cfg.Port` (from env, not literal) | INSTRUCTIONS §3 |
| `/health` handler | — | `{status:"ok", service:"proxy-validator", version}` | API_CONTRACTS.md "Health (all services)" |
| default pool key | — | `proxy_pool` | DATABASE_SCHEMA §3 |
| default pool TTL | — | `5*time.Minute` | DATABASE_SCHEMA §3 (5 min) |
| pool write | — | Redis Set via `SAdd` + `Expire` | DATABASE_SCHEMA §3 ("Redis Set") |
| pool replace | — | staging key + `Rename` (atomic) | A-04 resilience (no empty pool) |
| politeness limiter | — | randomised delay; `Failure()` backs off, `Success()` relaxes | A-04 task ("adaptive random delays") |
| runner dispatch | — | calls `limiter.Wait` and reports `Failure`/`Success` | A-04 task (politeness applied) |
| fetcher many sources | one fails | other sources still parsed; error only if **all** fail | NFR-2 (resilience) |
| default port | — | `8086` | ARCHITECTURE §2 / `.env.example` |
| each package file | — | begins with `// Package …` (main: `// Command …`) | INSTRUCTIONS §3 (GoDoc) |
