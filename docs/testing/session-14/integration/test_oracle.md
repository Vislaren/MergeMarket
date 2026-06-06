## Oracle — Integration (running binaries)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| search binary, Postgres unreachable | startup | non-zero exit, error logged | ARCHITECTURE §2 (Postgres hard dep) |
| search binary + live DB, `GET /search?q=&location=` | valid | 200 `{results, cached, latency_ms}` | API_CONTRACTS.md → Search |
| search, same query twice (Redis up) | cache | 1st `cached:false`, 2nd `cached:true`, faster | DATABASE_SCHEMA §3 + ARCH §10 |
| search, `q` omitted | invalid | 400 `missing_query` | API_CONTRACTS.md |
| userdata binary, no `JWT_SECRET` | startup | refuses to start (non-zero) | shares auth secret (A-08) / NFR-4 |
| userdata binary, Postgres unreachable | startup | non-zero exit, error logged | ARCHITECTURE §2 |
| userdata, `GET /wishlist` no token | unauth | 401 `unauthorized` | API_CONTRACTS.md (Authorization) |
| userdata, user B deletes user A's wishlist id | not owned | 404 `not_found` | per-user `user_id` scoping |
| full stack: search→wishlist→alert→drop | E2E | notification delivered for the followed product | USER_FLOWS Flow 4/5/6; session-13 TC-13-I |
