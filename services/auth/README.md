# Auth Service

Go service for MergeMarket registration, login, and refresh-token rotation.

## Endpoints

- `GET /health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

All API responses follow `project_docs/api/API_CONTRACTS.md`.

## Runtime Notes

- PostgreSQL stores users in the `users` table.
- Passwords are bcrypt hashes.
- Refresh sessions are stored in Redis under `session:{sha256(refresh_token)}` with a 1-hour TTL.
- Redis session payloads are encrypted with AES-256-GCM using `AUTH_ENCRYPTION_KEY`.
- The HTTP server requires TLS certificates and enforces TLS 1.3.

Required environment variables are listed in `.env.example`.
