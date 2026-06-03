## Test Plan — Integration — Session 02 — A-02 docker-compose for Local Dev

**Scope:** Verify the A-02 stack actually works end-to-end when started with
`docker compose up`:
- All five services reach a **healthy** state within their `start_period`.
- The bootstrap schema is applied to a fresh Postgres volume: all tables,
  the `job_status` enum, and the `price_history` **hypertable** exist, with
  the `timescaledb` extension enabled.
- Redis is reachable (PING→PONG) and honours `REDIS_PASSWORD`.
- Kong boots in DB-less mode, loads `kong.yml`, and exposes proxy + admin.
- SonarQube reports `status: UP` and is backed by `sonar-db` (not embedded H2).
- Named volumes persist data across a `down` (without `-v`) / `up` cycle.

**Out of scope:**
- Kong route behaviour (no routes exist until A-09).
- SonarQube quality-gate logic and the CI scan (A-10/A-11).
- Performance / NFR latency targets — no application services run yet.

**Approach:** Integration tests verify services working together over their
real interfaces (SQL connection, Redis protocol, HTTP) — **no mocks**. The
stack is provisioned with `docker compose up -d`; assertions then run against
the exposed localhost ports.

**Entry criteria:**
- A running Docker daemon.
- `.env` present (`cp .env.example .env`).
- Host satisfies SonarQube's `vm.max_map_count >= 262144` (Linux).

**Exit criteria:** All integration cases PASS; `docker compose ps` shows every
service `(healthy)`.

**Tools:** `docker compose`, `psql`/`pgx`, `redis-cli`/go-redis, `curl`/net-http,
Go `testing` + `testify` (suite tagged `//go:build integration`).

**Assumptions:**
- First boot uses an **empty** `pgdata` volume so `/docker-entrypoint-initdb.d`
  runs (the init script only executes on an uninitialised data dir).
- Image pulls succeed (timescaledb, redis, kong, sonarqube, postgres).

**Risk / status:**
- **Not executed this session — Docker daemon was not running.** All
  integration cases are therefore **PENDING**. They are written and ready to
  run once a daemon is available (locally or in the A-10 CI pipeline).
- SonarQube boot is slow (~1-2 min); flaky health timing is mitigated by a
  90s `start_period` and 10 retries.
