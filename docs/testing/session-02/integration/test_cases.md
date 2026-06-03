# Test Cases — Integration — Session 02 — A-02

All cases are **PENDING**: the Docker daemon was not running this session, so
the live stack could not be started. Each case below is ready to execute once
a daemon is available (or in CI / A-10).

---

### TC-02-I-001: Stack starts and all services become healthy
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Integration |
| Preconditions | Docker daemon up; `.env` present |
| Input | `docker compose up -d` |
| Steps | 1. `cp .env.example .env` 2. `docker compose up -d` 3. Poll `docker compose ps` until all healthy or timeout |
| Expected result | All 5 services report `(healthy)` within their start_period |
| Actual result | [PENDING] |
| Notes | sonarqube start_period 90s; others ≤20s. |

---

### TC-02-I-002: Bootstrap schema applied
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Integration |
| Preconditions | postgres healthy; fresh `pgdata` volume |
| Input | Connection to `localhost:5432` db `mergemarket` |
| Steps | 1. Connect as `${DB_USER}` 2. Query `information_schema.tables` |
| Expected result | All 8 tables exist: users, stores, products, wishlist_items, price_alerts, return_policies, scrape_jobs, price_history |
| Actual result | [PENDING] |
| Notes | Init script runs only on an empty data volume. |

---

### TC-02-I-003: TimescaleDB extension + hypertable
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Integration |
| Preconditions | postgres healthy |
| Input | SQL queries |
| Steps | 1. `SELECT extname FROM pg_extension WHERE extname='timescaledb'` 2. `SELECT * FROM timescaledb_information.hypertables WHERE hypertable_name='price_history'` |
| Expected result | Extension present; `price_history` listed as a hypertable |
| Actual result | [PENDING] |
| Notes | Validates the single-instance Postgres+Timescale decision behaviourally. |

---

### TC-02-I-004: Redis reachable and honours password setting
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Integration |
| Preconditions | redis healthy |
| Input | Redis PING on `localhost:6379` |
| Steps | 1. `redis-cli ${REDIS_PASSWORD:+-a $REDIS_PASSWORD} ping` |
| Expected result | `PONG`. With empty `REDIS_PASSWORD` auth is disabled; when set, unauthenticated PING is rejected |
| Actual result | [PENDING] |
| Notes | Verifies the conditional `--requirepass` entrypoint logic. |

---

### TC-02-I-005: Kong loads declarative config (DB-less)
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Integration |
| Preconditions | kong healthy |
| Input | `GET http://localhost:8001/` and `GET http://localhost:8001/services` |
| Steps | 1. Hit admin root 2. List services |
| Expected result | Admin responds 200; `database: "off"`; services list empty (placeholder kong.yml) |
| Actual result | [PENDING] |
| Notes | Confirms the placeholder loads cleanly; A-09 will populate routes. |

---

### TC-02-I-006: Kong proxy port listening
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Integration |
| Preconditions | kong healthy |
| Input | `GET http://localhost:8000/` |
| Steps | 1. Hit proxy port |
| Expected result | Connection accepted; Kong replies (404 "no Route matched" is expected with the empty config) |
| Actual result | [PENDING] |
| Notes | 404 here is a PASS — it proves the proxy is up, just has no routes yet. |

---

### TC-02-I-007: SonarQube UP and backed by sonar-db
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Integration |
| Preconditions | sonarqube healthy; sonar-db healthy |
| Input | `GET http://localhost:9000/api/system/status` |
| Steps | 1. Poll status endpoint 2. Inspect sonarqube logs for the postgres JDBC URL |
| Expected result | `{"status":"UP"}`; logs show `jdbc:postgresql://sonar-db:5432/sonar` (not H2) |
| Actual result | [PENDING] |
| Notes | Confirms the dedicated-DB decision over evaluation-only H2. |

---

### TC-02-I-008: Named volumes persist across restart
| Field | Value |
|-------|-------|
| Task reference | A-02 |
| Type | Integration |
| Preconditions | Stack up; a row written to `mergemarket` |
| Input | `docker compose down` (no `-v`) then `up -d` |
| Steps | 1. INSERT a test row 2. `down` 3. `up -d` 4. SELECT the row |
| Expected result | Row still present; schema not re-initialised (init skipped on existing volume) |
| Actual result | [PENDING] |
| Notes | Verifies `pgdata` persistence and idempotent init. |
