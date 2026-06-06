## Test Plan — Integration — Session 14 — A-14 (Search) + A-16..A-18 (User-data)

**Scope:** The real search (A-14) and userdata (A-16..A-18) **binaries** running
as processes: startup/dependency behaviour, and their HTTP contract against a
live Postgres/Redis.

**Out of scope:** Internal function logic (covered by the unit plan); the Flutter
client; the truth-score service (A-15).

**Approach:** The suite **builds and runs the actual binaries**. Two classes:
- **Executable here (no DB needed):** fail-fast / config-guard behaviour — a
  service must refuse to start without its required secret and must exit non-zero
  when a hard dependency (Postgres) is unreachable (NFR-2 resilience / fail-fast).
- **Gated (PENDING locally):** cases needing a live backend self-skip unless
  `DB_TEST_DSN` (and, for the cache case, `REDIS_TEST_ADDR`) are set.

**Entry criteria:** Go toolchain present; services build. For the gated cases: a
reachable Postgres with the `01-schema.sql` applied (and Redis for the cache
case).

**Exit criteria:** All executable cases PASS; gated cases PASS once a backend is
provided.

**Tools:** Go `testing` + `testify`, `//go:build integration` tag.

**Assumptions:** No Docker is available locally and the host Postgres on :5432 is
of unknown ownership/credentials, so the live cases are **not** run against it
(safety) — they are written, compile under `-tags=integration`, and skip cleanly.

**Risk / why PENDING:** The full search→results→wishlist→alert E2E (TC-14-I-008,
carrying forward session-13's TC-13-I) needs the **whole stack** up together
(Kong + BFF + auth + search + userdata + Postgres + Redis). That is a CI /
deployed-environment exercise, not a local one.

**How to run the gated cases:**
```bash
DB_TEST_DSN=postgres://postgres:pass@localhost:5432/mergemarket?sslmode=disable \
REDIS_TEST_ADDR=localhost:6379 \
go test -tags=integration ./docs/testing/session-14/integration/test_suite/...
```
