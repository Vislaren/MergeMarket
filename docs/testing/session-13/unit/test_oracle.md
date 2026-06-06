## Oracle — B-11 AuthenticatedClient / BFF JWT forwarding

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| Protected request, session present | token in `readToken` | request carries `Authorization: Bearer <token>` | API_CONTRACTS (Bearer JWT on all non-auth routes) + Kong JWT plugin (A-09) |
| Protected request, signed out | `readToken → null` | no `Authorization` header; request passes through unchanged | B-02 mock requires no auth; preserves prior tests |
| Protected request → 401, token sent | refresh succeeds | one refresh, original request replayed with new token, success returned | B-08 follow-up (refresh-on-401); OAuth refresh semantics |
| Protected request → 401, token sent | refresh fails | original 401 surfaces (no loop); session cleared → router routes to login | API_CONTRACTS 401; USER_FLOWS Flow 1 |
| `POST /api/v1/auth/*` → 401 | always | no refresh; 401 returned as-is | auth endpoints mint tokens — a 401 is a credential failure |
| N concurrent protected 401s | refresh in flight | exactly one refresh; all requests replayed | single-flight interceptor design |
| `refreshSession()` | refresh 200 | new session persisted + state updated, returns true | API_CONTRACTS auth/refresh 200 |
| `refreshSession()` | refresh 401 | session cleared, returns false | API_CONTRACTS auth/refresh 401 (token_expired) |
| `GET /products/{id}/detail` with `Authorization` | BFF aggregate | header forwarded to all upstream calls | A-09 Kong → BFF → upstream JWT gate |
| Default `API_BASE_URL` unset | app boot | base URL = `http://localhost:8089` (not blocked `:8080`) | PORTS_README (mock server = 8089) |
