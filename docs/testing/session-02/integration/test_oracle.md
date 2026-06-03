## Oracle — A-02 Local Dev Stack (live)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `docker compose up -d` | fresh env | all 5 services `(healthy)` | A-02 (health checks) |
| query `information_schema.tables` | fresh pgdata volume | 8 tables present | DATABASE_SCHEMA §1 |
| `pg_extension` lookup | postgres up | `timescaledb` enabled | DATABASE_SCHEMA §2 |
| `timescaledb_information.hypertables` | postgres up | `price_history` is a hypertable | DATABASE_SCHEMA §2 |
| `redis-cli ping` | empty REDIS_PASSWORD | `PONG`, no auth required | A-02 redis entrypoint |
| `redis-cli ping` (no auth) | REDIS_PASSWORD set | `NOAUTH`/auth error | A-02 redis entrypoint |
| `GET :8001/` | kong up | 200, `database:"off"` | ARCHITECTURE §2 / Kong DB-less |
| `GET :8001/services` | placeholder kong.yml | empty list | A-02 kong.yml stub |
| `GET :8000/` | empty route table | 404 "no Route matched" (proxy alive) | Kong behaviour |
| `GET :9000/api/system/status` | sonarqube booted | `status: UP` | A-02 / SonarQube API |
| sonarqube logs | startup | JDBC URL `sonar-db:5432/sonar` (not H2) | A-02 decision |
| `down` (no `-v`) then `up` | prior data written | data persists; init not re-run | A-02 named volumes + idempotent init |
