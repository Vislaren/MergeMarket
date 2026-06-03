# Auth Service (Go)

Issues and validates JWT tokens; manages user registration, login, and
token refresh. Sessions cached in Redis (TTL 1h). AES-256 at rest, TLS 1.3
in transit.

- **Port:** 8081
- **Endpoints:** `POST /api/v1/auth/{register,login,refresh}`, `GET /health`
- **Task:** A-08

_Implementation pending._
