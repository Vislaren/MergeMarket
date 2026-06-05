# Agent A — TODO

> Tasks are listed in the order they must be completed.
> Pick the first task that is not blocked.
> When a task is completed, remove it from this file and add it to DONE.md.
> When a task is blocked, move it to BLOCKED.md and pick the next one.

---

## Task Queue

---

### A-10 — Jenkins CI/CD Pipeline
**Description:** Write the `Jenkinsfile` at the repo root defining a
multibranch pipeline with four stages: Checkout, Test (go test across
all services), SonarQube Analysis (sonar-scanner), Quality Gate
(waitForQualityGate — aborts pipeline on failure), and Deploy (SSH into
VPS, git pull, docker compose up — runs on `main` branch only).
Jenkins is already installed on the VPS. Configure the pipeline job in
Jenkins as a Multibranch Pipeline pointing to the GitHub repository.
Set up the GitHub webhook at `http://95.111.228.35:8080/github-webhook/`
so every push triggers the pipeline automatically.

**Depends on:** A-02

**Output:** `Jenkinsfile` at repo root, `docs/DEPLOYMENT.md` updated

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
