// Module for the Session-02 test suites (unit + integration).
// Self-contained so the suites can be run independently of the (not-yet-created)
// service modules. Run `go mod tidy` to populate the dependency graph / go.sum
// in CI before `go test`.
module github.com/Vislaren/MergeMarket/docs/testing/session-02

go 1.22

require (
	github.com/jackc/pgx/v5 v5.6.0
	github.com/redis/go-redis/v9 v9.6.0
	github.com/stretchr/testify v1.9.0
)
