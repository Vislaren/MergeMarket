## Oracle — A-02 Local Dev Environment (static artefacts)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `docker compose config` | `.env` populated from `.env.example` | Exit 0, valid interpolated config | Compose spec / A-02 |
| `docker compose config --services` | — | {postgres, redis, kong, sonarqube, sonar-db} | A-02 task description + DONE decision |
| Service definition | each service | exactly one `healthcheck` block | A-02 task ("health checks") |
| `docker compose config --volumes` | — | 6 named volumes (pgdata, redisdata, sonar_db_data, sonarqube_data, sonarqube_extensions, sonarqube_logs) | A-02 task ("named volumes") |
| `ports` blocks | — | 5432, 6379, 8000/8443/8001/8444, 9000; sonar-db unpublished | ARCHITECTURE §2 (ports), A-02 |
| Go service ports vs host ports | — | no overlap with 8081-8086 | ARCHITECTURE §2 |
| `.env.example` keys | — | superset of DATABASE_SCHEMA §4 variables | DATABASE_SCHEMA §4 |
| `api-gateway/kong.yml` | DB-less mode | valid declarative doc, `_format_version` set | Kong DB-less docs / A-09 contract |
| `01-schema.sql` objects | — | 8 tables + job_status enum + price_history hypertable + timescaledb extension | DATABASE_SCHEMA §1-2 |
| DB_* vs TIMESCALE_* | — | identical host:port (single instance) | DATABASE_SCHEMA §4 + A-02 decision |
