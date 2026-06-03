# Test Cases — Unit — Session 02 — A-02

All cases below were executed this session via `docker compose config` and
shell assertions against the working tree at commit `ee844d0`.

---

### TC-02-U-001: Compose file parses
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | `.env` present (copied from `.env.example`); `docker compose` CLI installed |
| Input | `docker-compose.yml` + `.env` |
| Steps | 1. `cp .env.example .env` 2. `docker compose config` |
| Expected result | Exits 0; fully interpolated config emitted; no errors |
| Actual result | [PASS] |
| Notes | Daemon not required for `config`. |

---

### TC-02-U-002: Required infrastructure services defined
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | Compose parses (TC-02-U-001) |
| Input | `docker compose config --services` |
| Steps | 1. List services 2. Assert the set |
| Expected result | `postgres, redis, kong, sonarqube, sonar-db` all present |
| Actual result | [PASS] |
| Notes | `postgres` = timescale/timescaledb satisfies both "PostgreSQL" and "TimescaleDB" line items (A-02 decision). `sonar-db` is the dedicated backing DB for SonarQube. |

---

### TC-02-U-003: Every service declares a health check
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | Compose parses |
| Input | Resolved compose config |
| Steps | 1. Count `healthcheck.test` entries |
| Expected result | 5 health checks (one per service) |
| Actual result | [PASS] |
| Notes | postgres→`pg_isready`, redis→`redis-cli ping`, kong→`kong health`, sonar-db→`pg_isready`, sonarqube→`/api/system/status`. |

---

### TC-02-U-004: All named volumes declared
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | Compose parses |
| Input | `docker compose config --volumes` |
| Steps | 1. List volumes 2. Assert the set |
| Expected result | `pgdata, redisdata, sonar_db_data, sonarqube_data, sonarqube_extensions, sonarqube_logs` |
| Actual result | [PASS] |
| Notes | 6 named volumes — persistence for both databases and all SonarQube state. |

---

### TC-02-U-005: Port mappings correct and env-driven
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | Compose parses |
| Input | Resolved `ports:` blocks |
| Steps | 1. Inspect published→target ports |
| Expected result | postgres 5432, redis 6379, kong 8000/8443 (proxy) + 8001/8444 (admin), sonarqube 9000; sonar-db has **no** published port |
| Actual result | [PASS] |
| Notes | Host ports come from `.env` (`${DB_PORT}` etc.); none collide with the Go service ports 8081-8086. sonar-db correctly internal-only. |

---

### TC-02-U-006: `.env.example` contains all required variables
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | `.env.example` present |
| Input | `.env.example` |
| Steps | 1. Assert each key from `DATABASE_SCHEMA.md §4` + compose extras exists |
| Expected result | All present: DB_*, TIMESCALE_*, REDIS_*, JWT_*, REFRESH_TOKEN_EXPIRY_DAYS, all 6 service ports, SONAR_HOST_URL, SONAR_TOKEN, FIREBASE_SERVER_KEY, plus Kong port + SONAR_DB_* extras |
| Actual result | [PASS] |
| Notes | 23/23 required keys found; extras (`KONG_*_PORT`, `SONAR_DB_USER/PASSWORD`, `COMPOSE_PROJECT_NAME`) also present. |

---

### TC-02-U-007: Kong declarative config is valid
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | `api-gateway/kong.yml` present |
| Input | `api-gateway/kong.yml` |
| Steps | 1. Assert `_format_version` present and `services`/`routes` keys valid |
| Expected result | Valid DB-less declarative document; `_format_version: "3.0"` |
| Actual result | [PASS] |
| Notes | Placeholder by design — empty `services`/`routes`. A-09 replaces it. Behavioural load is verified in TC-02-I-005. |

---

### TC-02-U-008: Bootstrap schema declares all DB objects
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | `infra/db/init/01-schema.sql` present |
| Input | `01-schema.sql` |
| Steps | 1. Assert each object is declared |
| Expected result | `timescaledb` extension; tables users, stores, products, wishlist_items, price_alerts, return_policies, scrape_jobs, price_history; `job_status` enum; `create_hypertable(price_history)` |
| Actual result | [PASS] |
| Notes | All 8 relational tables + hypertable + extension present and match `DATABASE_SCHEMA.md`. Syntactic apply against a live PG is TC-02-I-002. |

---

### TC-02-U-009: Single Postgres+Timescale connection (design contract)
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Unit |
| Preconditions | `.env.example` present |
| Input | DB_HOST/DB_PORT vs TIMESCALE_HOST/TIMESCALE_PORT |
| Steps | 1. Assert both pairs resolve to the same host:port |
| Expected result | `DB_HOST==TIMESCALE_HOST` and `DB_PORT==TIMESCALE_PORT` (both `localhost:5432`) |
| Actual result | [PASS] |
| Notes | Confirms the A-02 "one instance serves relational + hypertable" decision matches `DATABASE_SCHEMA.md §4`. |
