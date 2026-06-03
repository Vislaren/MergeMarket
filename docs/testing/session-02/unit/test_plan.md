## Test Plan — Unit — Session 02 — A-02 docker-compose for Local Dev

**Scope:** Static validation of each artefact A-02 produced, treated as an
isolated unit, with **no containers running**:
- `docker-compose.yml` — parses; declares the required services, a health
  check per service, named volumes, and correct env-driven port mappings.
- `.env.example` — contains every variable required by `DATABASE_SCHEMA.md §4`
  plus the compose-specific extras (Kong ports, SonarQube DB credentials).
- `api-gateway/kong.yml` — is a valid DB-less declarative document.
- `infra/db/init/01-schema.sql` — declares all relational tables, the
  `job_status` enum, the TimescaleDB extension, and the `price_history`
  hypertable, matching `DATABASE_SCHEMA.md`.

**Out of scope:**
- Actually starting containers, applying the schema, or hitting endpoints —
  that is the integration suite for this session.
- Application/business logic — A-02 ships no Go or Dart code.
- The full Kong route table / JWT plugin — `kong.yml` is a placeholder owned
  by A-09; only its validity as a declarative document is checked here.

**Approach:** Unit tests isolate individual artefacts and assert their
structure by parsing/reading the file. `docker compose config` is used as the
authoritative parser for the Compose file (it does not require the daemon).
No mocks are needed — the artefacts are static files.

**Entry criteria:**
- A-02 committed (`ee844d0`) and the four artefacts present in the working tree.
- `docker compose` CLI available (daemon not required for `config`).

**Exit criteria:** All unit test cases PASS. (No coverage % is meaningful for
config artefacts; completeness is asserted by enumerated key/object checks.)

**Tools:** `docker compose config`, shell assertions (executed this session);
Go `testing` + `testify` for the committed CI-runnable suite.

**Assumptions:**
- `timescale/timescaledb` image provides PostgreSQL with the TimescaleDB
  extension on a single connection (A-02 decision) — verified structurally
  here, behaviourally in the integration suite.
- The Go suite runs in CI (A-10) where module downloads are available; it was
  not executed locally this session (no module cache fetched), but the
  equivalent static assertions were executed via shell and all passed.

**Risk:**
- Static checks confirm the artefacts are *well-formed and complete*; they
  cannot confirm the stack actually *boots healthy* — that gap is covered
  (PENDING) by the integration suite until a Docker daemon is available.
- `latest-pg16` is a floating tag; a future image change could alter runtime
  behaviour without any change to these files.
