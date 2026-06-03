# Agent A — TODO

> Tasks are listed in the order they must be completed.
> Pick the first task that is not blocked.
> When a task is completed, remove it from this file and add it to DONE.md.
> When a task is blocked, move it to BLOCKED.md and pick the next one.

---

## Task Queue

---

### A-02 — docker-compose.yml for Local Dev
**Description:** Write the `docker-compose.yml` at the project root that
spins up PostgreSQL, TimescaleDB, Redis, Kong, and SonarQube as local
containers. Include named volumes, health checks, and correct port mappings.
Add a `.env.example` file with all required environment variables.

**Depends on:** A-01

**Output:** `docker-compose.yml`, `.env.example`

---

### A-04 — Proxy-Validator Service
**Description:** Build the Go service that scrapes public proxy lists,
tests each proxy for viability against a real endpoint, and writes the
working IP list to Redis with a 5-minute TTL. Implement the politeness
protocol (adaptive random delays).

**Depends on:** A-02

**Output:** `services/proxy-validator/` (complete Go service)

---

### A-05 — Scraper Service
**Description:** Build the config-driven worker-queue scraper engine.
Implement the `StoreConfig` JSON/YAML loader. Workers consume jobs from
the queue, load the relevant config, execute the scrape, and write raw
results to a normalisation queue. Implement the Circuit Breaker pattern
for consecutive 403/429 failures.

**Depends on:** A-04

**Output:** `services/scraper-service/` including sample `configs/` files

---

### A-06 — Normalization Service
**Description:** Build the Go service that consumes raw scrape output and
converts it into the standard product schema (Price, Currency, Title,
Image URL, Shipping). Implement the Affiliate Link Injection module that
wraps outbound URLs with retailer-specific parameters.

**Depends on:** A-05

**Output:** `services/normalization/`

---

### A-07 — History Service
**Description:** Build the Go service that records daily price snapshots
to TimescaleDB. Implement the scheduled heartbeat that scrapes followed
product URLs and triggers price-drop alert logic when a threshold is met.

**Depends on:** A-06

**Output:** `services/history/`

---

### A-08 — Auth Service
**Description:** Build the Go JWT service. Implement register, login, and
token refresh endpoints. Use AES-256 for data at rest and enforce TLS 1.3.
Store sessions in Redis with a 1-hour TTL.

**Depends on:** A-02

**Output:** `services/auth/`

---

### A-09 — Kong API Gateway Configuration
**Description:** Write Kong declarative configuration (`kong.yml`) that
defines all routes, rate limiting, and the JWT auth plugin. Map all API
contract endpoints from INSTRUCTIONS.md §4.

**Depends on:** A-08

**Output:** `api-gateway/kong.yml`, `api-gateway/plugins/`

---

### A-10 — GitHub Actions CI/CD Pipeline
**Description:** Write the `.github/workflows/ci.yml` pipeline that runs
`go test` with coverage on all services, executes `sonar-scanner` against
the VPS SonarQube instance, and enforces the quality gate. Pipeline must
fail if the quality gate does not pass.

**Depends on:** A-02

**Output:** `.github/workflows/ci.yml`

---

### A-11 — SonarQube on VPS
**Description:** Deploy SonarQube Community Edition as a persistent Docker
container on the VPS with named volumes for data and logs. Open port 9000
on the firewall. Document how to generate the CI token.

**Depends on:** A-03

**Output:** `infra/k3s/sonarqube.yml`, `docs/sonarqube-setup.md`

---

### A-12 — Grafana Dashboard on VPS
**Description:** Deploy Grafana as a Docker container on the VPS (port
3000). Configure two data sources: SonarQube API and GitHub API. Build a
dashboard with panels for test coverage %, bug count, vulnerability count,
pipeline pass/fail rate, and build duration over time.

**Depends on:** A-11

**Output:** `infra/k3s/grafana.yml`, `infra/grafana/dashboards/`

---

### A-13 — Automated Database Backups
**Description:** Write a cron-based backup script that runs nightly on the
VPS. The script must: dump the full PostgreSQL/TimescaleDB database using
`pg_dump`, compress the output, and sync it to free external storage.
Use Backblaze B2 (free tier: 10GB) via the `rclone` tool for the sync.
Store the last 7 daily backups and delete older ones automatically.
Add the cron entry to a systemd timer so it survives VPS reboots.
Document the restore procedure in `docs/backup-restore.md`.

**Depends on:** A-02

**Output:** `infra/scripts/backup.sh`, `infra/scripts/restore.sh`,
`infra/backup.timer`, `infra/backup.service`, `docs/backup-restore.md`
