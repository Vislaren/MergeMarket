# Mock Server (Go) — local dev only, not deployed

Serves every endpoint in [`project_docs/api/API_CONTRACTS.md`](../../project_docs/api/API_CONTRACTS.md)
with static fixtures so the Flutter app and BFF can be developed offline before
Agent A's real services exist. **Dependency-free** (Go standard library only) so
it always builds and runs without a network round-trip.

- **Task:** B-02
- **Port:** `8080` (override with `MOCK_SERVER_PORT` or `PORT`)
- **Endpoints:** every `/api/v1/*` contract + `GET /health`
- **CORS:** permissive (`*`) with preflight handling, for Flutter web/dev clients

## Run

```bash
cd services/mock-server
go run ./cmd/mock-server                 # listens on :8080
MOCK_SERVER_PORT=8091 go run ./cmd/mock-server   # if 8080 is taken (e.g. Jenkins)
```

> Note: a local Jenkins also defaults to port 8080. If you run both, start the
> mock server on another port via `MOCK_SERVER_PORT`.

## Test

```bash
go build ./...
go test ./... -cover
```

## Layout

```
services/mock-server/
├── cmd/mock-server/main.go     # entrypoint, graceful shutdown
└── internal/
    ├── config/                 # env-driven config (port, identity)
    ├── fixtures/               # response types + static sample data
    └── server/                 # routing (Go 1.22 patterns), handlers, CORS
```

## Sentinel inputs (exercise non-200 paths without server state)

The server is stateless; deterministic sentinel inputs drive the error paths so
clients can test them:

| Endpoint | Sentinel | Result |
|----------|----------|--------|
| `GET /api/v1/search` | `q` empty | `400 missing_query` |
| `GET /api/v1/search` | `q=timeout` | `504 timeout` |
| `POST /api/v1/auth/register` | missing email/password | `400 invalid_input` |
| `POST /api/v1/auth/register` | email `taken@mergemarket.app` | `409 email_exists` |
| `POST /api/v1/auth/login` | password `wrongpassword` | `401 invalid_credentials` |
| `POST /api/v1/auth/refresh` | refresh_token `expired` | `401 token_expired` |
| `GET /api/v1/products/{id}/history` | id `unknown` | `404 not_found` |
| `GET /api/v1/products/{id}/truth-score` | id `unknown` | `404 not_found` |
| `POST /api/v1/wishlist` | product_id `prod-001` | `409 already_in_wishlist` |
| `POST /api/v1/alerts` | missing product_id / threshold ≤ 0 | `400 invalid_input` |
| `DELETE /api/v1/{wishlist,alerts}/{id}` | id `unknown` | `404 not_found` |

All other valid requests return the documented success response with sample data.
