# Agent A — Done

> This file is the permanent record of everything Agent A has completed.
> Add one entry per completed task. Never delete entries.
> Agent B reads this file to know what to test each session.

---

## Completed Tasks

---

### [DONE] A-03 — VPS Setup
**Session:** Pre-project (completed manually before agent sessions began)
**Completed by:** User (manual setup)
**Commit:** N/A — done outside agent workflow

**What was built:**
VPS fully set up with Docker, K3s, Helm, and Terraform installed on the
Ubuntu 22.04 Contabo VPS. K3s is running and the cluster node is Ready.

**Files created/modified:**
- Manual setup — no files committed yet. Agent A should document this
  in `docs/vps-setup.md` during session A-01 or A-02 for reference.

**Key decisions made:**
- Using Contabo VPS with Ubuntu 22.04
- K3s chosen over full Kubernetes for single-node operation
- Docker installed for container runtime

**API contracts changed:** No

**Known limitations:**
- Setup not yet documented in the repo. Should be captured in
  `docs/vps-setup.md` during an early session.

---

### [DONE] A-01 — Initialise Repo Structure
**Session:** 1
**Completed by:** Agent A
**Commit:** `session(A-01): initialise repo structure` _(local commit — push deferred, see limitations)_

**What was built:**
The full MergeMarket folder layout as defined in
`project_docs/architecture/ARCHITECTURE.md §5` (the doc referenced in
INSTRUCTIONS as `Documentation.md §6.2` does not exist — §5 of ARCHITECTURE
is the authoritative structure). Added a root `.gitignore` covering Go,
Flutter, Terraform, SonarQube, and environment files. Added descriptive
`README.md` files in each service folder and `.gitkeep` placeholders in
every otherwise-empty directory so Git tracks the full tree. Also captured
the manual A-03 VPS setup in `docs/vps-setup.md` (per the DONE.md note on
the A-03 entry) and added a root project `README.md`.

**Files created/modified:**
- `.gitignore`
- `README.md`
- `docs/vps-setup.md`
- `api-gateway/README.md`, `api-gateway/plugins/.gitkeep`
- `services/{auth,bff,scraper-service,normalization,history,proxy-validator,mock-server}/README.md`
- `services/scraper-service/configs/.gitkeep`
- `apps/mobile/lib/{screens,widgets,providers,services,models,theme}/.gitkeep`
- `apps/mobile/test/{unit,mocks}/.gitkeep`, `apps/mobile/integration_test/.gitkeep`
- `infra/{k3s,terraform}/.gitkeep`, `shared/.gitkeep`, `docs/testing/.gitkeep`

**Key decisions made:**
- Used `.gitkeep` for empty leaf directories (Git does not track empty
  folders); service folders got real one-paragraph READMEs instead of empty
  files, which is more useful while still satisfying trackability.
- `.gitignore` whitelists `.env.example` and `*.tfvars.example` so the
  example files (A-02) remain committable while real secrets are ignored.
- Treated `project_docs/architecture/ARCHITECTURE.md §5` as the source of
  truth for the folder layout since `Documentation.md` is absent.

**API contracts changed:** No

**Known limitations / follow-ups for the user:**
- **Push deferred — repo not pushable yet.** Two blockers: (1) no Git
  remote is configured and no repository URL was provided (INSTRUCTIONS §7
  expected one at the first session); (2) the Git root is `C:/project`, a
  shared monorepo also containing unrelated projects (SafeZone, E_commerce,
  etc.), **not** a dedicated MergeMarket repo. Decide whether to (a) create
  a dedicated repo rooted at `C:/project/mergemarket`, or (b) push the
  shared repo to a provided remote. Then the §7 push protocol can run.
- A draft `.github/workflows/ci.yml` already exists at
  `mergemarket/.github/workflows/`. GitHub Actions only reads workflows at
  the **repo root**, so from `C:/project` this draft will not trigger. This
  is A-10's concern; flagged here so it isn't lost.

---

### [DONE] A-02 — docker-compose.yml for Local Dev
**Session:** 2
**Completed by:** Agent A
**Commit:** `session(A-02): docker-compose local dev environment`

**What was built:**
A root `docker-compose.yml` that brings up the full local backing
infrastructure, plus a matching `.env.example` and a turnkey DB bootstrap.

Services defined (all with health checks, named volumes, and env-driven
port mappings):
- `postgres` — `timescale/timescaledb:latest-pg16`. A **single** instance
  serves both the relational tables and the `price_history` hypertable
  (see decision below). Health check: `pg_isready`.
- `redis` — `redis:7-alpine`. Optional password via `REDIS_PASSWORD`
  (auth enabled only when the var is non-empty). Health check: `redis-cli ping`.
- `kong` — `kong:3.7` in **DB-less mode**, loading `api-gateway/kong.yml`.
  Proxy (8000/8443) and admin (8001/8444) ports exposed. Health: `kong health`.
- `sonarqube` — `sonarqube:community`, backed by a dedicated `sonar-db`,
  on port 9000. Health check polls `/api/system/status`.
- `sonar-db` — `postgres:16-alpine`, **not** exposed to the host; only
  SonarQube reaches it over the compose network.

Supporting files:
- `.env.example` — every required variable (DB, Timescale, Redis, JWT,
  service ports, Kong ports, SonarQube + its DB, Firebase), aligned with
  `DATABASE_SCHEMA.md §4` plus the compose-specific extras.
- `infra/db/init/01-schema.sql` — runs once on first boot of an empty
  Postgres volume; creates the timescaledb/pgcrypto extensions, all
  relational tables, indexes, the `job_status` enum, and the
  `price_history` hypertable. Mirrors `DATABASE_SCHEMA.md`.
- `api-gateway/kong.yml` — minimal valid DB-less placeholder so Kong boots
  cleanly now; **A-09 replaces it** with the full route/plugin config.

Validated with `docker compose config` (parses cleanly with a populated `.env`).

**Files created/modified:**
- `docker-compose.yml`
- `.env.example`
- `infra/db/init/01-schema.sql`
- `api-gateway/kong.yml` (placeholder — flagged for A-09)

**Key decisions made:**
- **One Postgres+Timescale container, not two.** The `timescale/timescaledb`
  image *is* PostgreSQL with the extension preloaded. `DATABASE_SCHEMA.md §4`
  points `DB_*` and `TIMESCALE_*` at the same `localhost:5432`, and the doc
  says "same connection as Postgres if using TimescaleDB extension". So the
  task's "PostgreSQL" and "TimescaleDB" line items are satisfied by this one
  service. This is simpler and matches the schema doc and the env contract.
- **Kong runs DB-less.** A-09's deliverable is a declarative `kong.yml`, which
  implies DB-less mode — so no Kong-specific database is needed.
- **SonarQube gets a dedicated `sonar-db`** rather than the embedded H2
  engine. H2 is evaluation-only and not persistent; a real Postgres with a
  named volume is the correct, stable setup.
- **Go services are not containerised here.** A-02 is the *infra* for local
  dev; services are run on the host and connect via the exposed localhost ports.

**API contracts changed:** No

**Known limitations / follow-ups:**
- **`api-gateway/kong.yml` is a placeholder.** A-09 must overwrite it with
  the real routes, rate limiting, and JWT plugin.
- SonarQube's embedded Elasticsearch needs host `vm.max_map_count >= 262144`
  on Linux — documented in the compose header.
- Push protocol still blocked by the same A-01 issue (no Git remote, shared
  `C:/project` root). Commit made locally; push deferred.

---

### [DONE] A-04 — Proxy-Validator Service
**Session:** 3
**Completed by:** Agent A
**Commit:** `session(A-04): proxy-validator service`

**What was built:**
The complete Go proxy-validator service — the **first real Go service** in the
repo, so it also establishes the module/layout convention for A-05..A-08.

It runs a continuous loop: scrape public proxy lists → parse & de-dup →
validate each proxy against a real endpoint (bounded concurrency, with an
adaptive politeness delay between dispatches) → write the working `ip:port`
set to Redis `proxy_pool` with a 5-minute TTL. Serves `GET /health` (API
contract shape) and `GET /stats` (last-cycle observability) on port 8086.

Pipeline packages (all with GoDoc, `log/slog`, env-only config, named errors):
- `internal/config` — env-driven `Config` + validation (rejects e.g.
  refresh ≥ TTL, max < min politeness).
- `internal/proxy` — `Addr` value type + tolerant proxy-list line parsing
  (strips scheme/whitespace/CR, drops comments, de-dups). 100% covered.
- `internal/fetcher` — downloads each source independently; one failing
  source never aborts the others (NFR-2); errors only if *all* sources fail.
- `internal/validator` — routes a real request through the proxy to the test
  URL; injectable `ClientFactory` makes it unit-testable without live proxies.
- `internal/politeness` — adaptive random-delay limiter: backs off on failure,
  relaxes on success, capped; context-aware `Wait`.
- `internal/store` — Redis persistence; **atomic pool swap** via staging key +
  `RENAME` so readers never see a half-written pool; empty result clears the key.
- `internal/runner` — orchestrates the cycle, bounded concurrency, tracks
  `Stats`, resilient `Loop` (cycle errors logged, never stop the loop).
- `internal/server` — `/health` + `/stats`, graceful shutdown.
- `cmd/proxy-validator/main.go` — wiring, SIGINT/SIGTERM graceful shutdown,
  tolerates Redis being down at startup (warns, retries next cycle).

**Verification:** `go build ./...`, `go vet ./...`, `gofmt -l` all clean;
`go test ./...` green (per-package coverage 78–100%, store's live-Redis tests
skip without `REDIS_TEST_ADDR`). Ran the binary: `/health` returns
`{"status":"ok","service":"proxy-validator","version":"0.1.0"}`, `/stats`
works, slog JSON logging confirmed, graceful handling of unreachable Redis/sources.

**Files created/modified:**
- `services/proxy-validator/go.mod`, `go.sum`
- `services/proxy-validator/cmd/proxy-validator/main.go`
- `services/proxy-validator/internal/{config,proxy,fetcher,validator,politeness,store,runner,server}/*.go` (+ `_test.go` each)
- `services/proxy-validator/Dockerfile`
- `services/proxy-validator/README.md` (replaced the stub)
- `.env.example` (added the Proxy Validator block + `REDIS_DB`)

**Key decisions made:**
- **Module per service:** `github.com/Vislaren/MergeMarket/services/proxy-validator`,
  `go 1.22`, Redis via `go-redis/v9` — matches the session-02 test module and
  INSTRUCTIONS §3. A-05..A-08 should follow the same `cmd/ + internal/` layout.
- **`proxy_pool` is a Redis Set** (per DATABASE_SCHEMA §3), written via an
  atomic staging-key `RENAME` so the pool is never transiently empty.
- **Politeness = adaptive random delay** (not a fixed sleep): randomised within
  a window, scaled by a failure/success-driven factor — satisfies the task's
  "adaptive random delays" requirement.
- **Default test URL** `http://www.google.com/generate_204` (fast 204); all of
  sources/URL/timeouts/concurrency are env-overridable.
- **Resilience first:** unreachable Redis or failing sources degrade gracefully
  rather than crash the service (NFR-2).

**API contracts changed:** No. (`/health` matches API_CONTRACTS.md exactly; the
extra `/stats` route is additive and not part of the public contract — Agent B
needs no mock changes.)

**Known limitations / follow-ups:**
- **SonarQube quality gate (§6) not verifiable this session:** the VPS
  (`95.111.228.35:9000` / `:8080`) is unreachable from the local dev
  environment, so the last pipeline's gate status couldn't be checked. No
  `[HOTFIX]` was needed locally (build/vet/fmt/tests all clean). Re-verify once
  CI runs against the pushed branch.
- **CI mismatch to resolve in A-10:** the working tree has the GitHub Actions
  `.github/workflows/ci.yml` deleted (dir removed) while a `Jenkinsfile` was
  committed in `cf991a3` — a half-finished Actions→Jenkins migration left by a
  prior session. I deliberately kept that limbo **out of this A-04 commit**
  (committed explicit paths, not `git add .`). A-10 owns reconciling which CI
  system is authoritative; ARCHITECTURE §3 still says "GitHub Actions" but
  INSTRUCTIONS §3 / TODO A-10 mandate Jenkins.
- Live free proxies are inherently flaky; expect a low working-ratio in real
  runs. Real-world tuning of sources/concurrency belongs to ops, not code.

**Git remote note:** A-01/A-02 flagged "no remote / shared C:/project root" as
a push blocker — **resolved now.** Repo root is the dedicated
`C:/project/mergemarket` and `origin` →
`https://github.com/Vislaren/MergeMarket.git`. `co-courage` is local-only and
is created on first push.

---

### [DONE] A-05 — Scraper Service
**Session:** 4 (built in a prior session; **DONE entry + commit were missing** —
reconstructed and committed this session alongside A-06/A-07)
**Completed by:** Agent A
**Commit:** `session(A-06..A-07): normalization + history services` _(A-05 files were
untracked on disk; they are committed together with A-06/A-07 this session)_

**What was built:**
The config-driven worker-queue scraper engine (`services/scraper-service/`,
module `github.com/Vislaren/MergeMarket/services/scraper-service`). Workers
`BLPOP` jobs from the `scrape_queue` Redis list, load the relevant per-store
`StoreConfig` (JSON/YAML from `configs/`), consult a per-store circuit breaker,
execute the scrape (optionally through a `proxy_pool` proxy from A-04), extract
products from the store's JSON API response, and `RPUSH` a `RawResult` onto the
`normalize_queue` for A-06. Serves `GET /health` (contract shape) + `GET /stats`
on port **8083**.

Packages: `config`, `storeconfig` (schema + dir loader → `Registry`),
`circuitbreaker` (per-store closed→open→half-open on consecutive 403/429),
`queue` (`Job`/`RawResult` types, `Queue`/`Sink` interfaces, Redis + in-memory),
`proxypool` (random proxy from the A-04 Set), `scraper` (URL render, request,
JSON field-path extraction), `worker` (pool: dequeue→breaker→scrape→publish),
`server` (`/health`+`/stats`). Sample configs: `example-jumia.json`,
`example-dummyjson.yaml`. Plus `Dockerfile` and README.

**Files created/modified:**
- `services/scraper-service/` — full tree (`cmd/`, `internal/{config,storeconfig,circuitbreaker,queue,proxypool,scraper,worker,server}` each with `_test.go`, `go.mod`, `go.sum`, `Dockerfile`, `README.md`, `configs/`).
- `.env.example` — Scraper service block (already present in the committed file).

**Key decisions made:**
- `mode: json_api` is implemented (preferred per ARCHITECTURE §2);
  `mode: html` (CSS selectors) is declared in the schema but extraction is not
  yet implemented (`scraper.ErrUnsupportedMode`).
- Circuit breaker state is currently in-memory (the `circuit:{store_id}` Redis
  key from DATABASE_SCHEMA §3 is reserved for a future shared-state version).
- Adding a store is a config file, not code (config-driven worker-queue pattern).

**API contracts changed:** No. `/health` matches the contract; `/stats` is an
additive observability route (no mock change needed for Agent B). It establishes
the `normalize_queue` **wire contract** (`RawResult` JSON) that A-06 consumes.

**Known limitations:**
- HTML/CSS-selector extraction not implemented yet.
- Breaker state not yet persisted to Redis (in-memory per process).

---

### [DONE] A-06 — Normalization Service
**Session:** 5
**Completed by:** Agent A
**Commit:** `session(A-06..A-07): normalization + history services`

**What was built:**
The Go normalization service (`services/normalization/`, module
`github.com/Vislaren/MergeMarket/services/normalization`, port **8084**). It
`BLPOP`s `RawResult` items off the `normalize_queue` Redis list (the A-05 wire
contract), converts each product to the canonical MergeMarket schema, injects
retailer-specific **affiliate links** (FR-6), and upserts the results into
PostgreSQL. Serves `GET /health` (contract shape) + `GET /stats`.

Pipeline: `queue` (BLPOP source, in-memory + Redis) → `normalize` (pure
transform: collapse/trim title, drop items with no title or non-positive price
[NFR-2], clamp negative shipping, ISO-4217 currency with USD fallback, compute
`total_cost = price + shipping`, round to 2dp) → `affiliate` (data-driven
injector) → `store` (pgx upsert). Packages: `config`, `queue`, `normalize`,
`affiliate`, `store`, `worker`, `server`, plus `cmd/normalization/main.go`,
`Dockerfile`, README, and `configs/affiliates.example.json`.

**Files created/modified:**
- `services/normalization/go.mod`, `go.sum`
- `services/normalization/cmd/normalization/main.go`
- `services/normalization/internal/{config,queue,normalize,affiliate,store,worker,server}/*.go` (+ `_test.go`)
- `services/normalization/Dockerfile`, `README.md`, `configs/affiliates.example.json`
- `.env.example` — Normalization service block
- `infra/db/init/01-schema.sql` — added `UNIQUE INDEX idx_products_store_url ON products(store_id, url)` (see API/schema note below)

**Key decisions made:**
- **Product upsert key = `(store_id, url)`.** A re-scrape updates the existing
  `products` row (price/shipping/affiliate/title/image/scraped_at) instead of
  duplicating. Required a `UNIQUE (store_id, url)` index — added to the canonical
  schema **and** created idempotently at startup (`EnsureSchema`) so the service
  is self-sufficient against an already-initialised DB.
- **Store resolution by name.** Each scraped store is upserted into `stores` by
  its unique `name` (`config_path` = scraper store id, `base_url` derived from
  the product URL host); resolved UUIDs are cached in memory to avoid re-querying
  on a hot stream.
- **Affiliate injection is data-driven** via a JSON config
  (`NORMALIZATION_AFFILIATE_CONFIG`): per-store `template` (deep-link wrapper with
  `{url}`/`{url_raw}` placeholders) and/or `params` (appended query params), with
  `default_params` for un-configured stores. A configured store fully specifies
  its own behaviour. No config = links pass through unchanged. Adding an affiliate
  program is a config change, not code.
- **PostgreSQL is a hard dependency** (it is the sink) — fail fast if unreachable.
  Redis being momentarily down at startup is tolerated (workers retry, NFR-2).
- **pgx/v5 (`pgxpool`)** chosen as the Postgres driver (first DB-touching service).
  A-07/A-08 follow the same driver/`Repository`-interface convention.

**API contracts changed:** No public HTTP contract change. **Schema note for
Agent B / future agents:** added `UNIQUE INDEX idx_products_store_url` to
`infra/db/init/01-schema.sql` — additive, only relevant to anyone relying on
products de-duplication. The canonical search-result fields are produced exactly
per API_CONTRACTS.md (minus `deal_score`, assigned downstream by the Deal Meter).

**Known limitations:**
- Live DB path (`internal/store`) is covered by an opt-in test gated on
  `DB_TEST_DSN` (skips without it), mirroring A-04's live-Redis gating. Unit runs
  stay green without a database.
- Normalized products are persisted to Postgres only; this service does not write
  the `search:{query_hash}` cache (that is the BFF/search concern, B-09).

---

### [DONE] A-07 — History Service
**Session:** 5
**Completed by:** Agent A
**Commit:** `session(A-06..A-07): normalization + history services`

**What was built:**
The Go history service (`services/history/`, module
`github.com/Vislaren/MergeMarket/services/history`, port **8085**). Three
responsibilities:
1. **Daily snapshots** — a scheduled loop (`HISTORY_SNAPSHOT_INTERVAL`, default
   24h) records a `price_history` row for every priced product in one bulk
   `INSERT … SELECT` into the TimescaleDB hypertable.
2. **Heartbeat** — a scheduled loop (`HISTORY_HEARTBEAT_INTERVAL`, default 1h)
   re-checks every **followed** product (those with an active `price_alerts`
   row), records a snapshot, and fires a **price-drop alert** on a *downward
   threshold crossing*.
3. **Price-history API** — `GET /api/v1/products/{product_id}/history` returning
   the exact API_CONTRACTS.md shape (`history[]`, `average_6m`, `lowest_30d`),
   plus `GET /health` and `GET /stats`.

Packages: `config`, `store` (pgx `Repository`: `SnapshotAll`,
`FollowedProducts`, `LatestPrice`, `InsertSnapshot`, `History`), `pricesource`
(`Source` interface; `DBSource` default + best-effort `HTTPSource`), `alert`
(`Event` + Redis `Publisher`), `runner` (snapshot + heartbeat loops, crossing
logic, stats), `server`. Plus `cmd/history/main.go`, `Dockerfile`, README.

**Files created/modified:**
- `services/history/go.mod`, `go.sum`
- `services/history/cmd/history/main.go`
- `services/history/internal/{config,store,pricesource,alert,runner,server}/*.go` (+ `_test.go`)
- `services/history/Dockerfile`, `README.md`
- `.env.example` — History service block

**Key decisions made:**
- **Alert fires on the downward crossing only:** current price ≤ threshold AND
  previous recorded price > threshold (or no prior snapshot). This fires once on
  the drop rather than every heartbeat while the price stays low. The previous
  price is captured from `price_history` *before* the new snapshot is inserted.
- **The service does not deactivate alerts.** It emits an `Event` to the Redis
  list `price_alert_events` (`HISTORY_ALERT_QUEUE`) for the Notification Worker
  (ARCHITECTURE §1); alert lifecycle/state stays with the alerts owner.
- **Heartbeat price source is pluggable** (`HISTORY_HEARTBEAT_MODE`): `db`
  (default — read `products.last_price`, reliable, no extra load) or `http`
  (re-fetch the product URL and best-effort extract an embedded price from
  JSON-LD / `product:price:amount` / `itemprop="price"`). HTTP honours the
  "scrape followed product URLs" wording but is inherently fragile, so it is
  opt-in; an unreadable page is skipped (NFR-2), not fatal.
- **History API computes aggregates in SQL** (`AVG` over 6 months, `MIN` over 30
  days via `FILTER`); `history[]` covers the last 6 months. Unknown product →
  `404 not_found` per the contract.
- Uses the Go 1.22 method+path routing (`GET /api/v1/products/{product_id}/history`,
  `req.PathValue`). Same pgx/`Repository` convention as A-06.

**API contracts changed:** No — implements `GET /api/v1/products/{id}/history`
exactly as specified (the response already matched the contract's
`average_6m`/`lowest_30d` fields). Adds an internal `price_alert_events` Redis
queue (new producer→Notification Worker contract; not a client-facing HTTP API).

**Known limitations:**
- Daily snapshot fires on an interval ticker (not a wall-clock "midnight"
  scheduler). Set `HISTORY_SNAPSHOT_ON_START=true` to take one immediately on
  boot. A precise time-of-day schedule can be added later (or via the A-13 cron).
- `http` heartbeat mode does direct fetches (no proxy rotation yet) and best-effort
  price extraction; the `db` default is recommended until per-store detail-page
  parsing exists.
- Live DB path covered by an opt-in `DB_TEST_DSN`-gated test (skips without it).

---

---

### [DONE] A-08 — Auth Service
**Session:** 6
**Completed by:** Agent A
**Commit:** `session(A-08..A-11): auth gateway sonarqube`

**What was built:**
The Go auth service (`services/auth/`, port **8091**) implementing
`POST /api/v1/auth/register`, `POST /api/v1/auth/login`,
`POST /api/v1/auth/refresh`, and `GET /health`. It stores users in PostgreSQL,
hashes passwords with bcrypt, issues HS256 JWT access tokens, rotates opaque
refresh tokens, and stores encrypted Redis session payloads with a strict
1-hour TTL.

Packages: `config` (env-only config + TLS/session validation), `store`
(pgx user repository), `token` (standard-library HS256 JWT issuer), `session`
(Redis `session:{sha256(refresh_token)}` store with AES-256-GCM encrypted
payloads), `secure` (AES-GCM + digest helpers), `service` (register/login/refresh
logic), `server` (TLS 1.3 HTTP API), and `cmd/auth/main.go` wiring.

**Files created/modified:**
- `services/auth/go.mod`, `go.sum`
- `services/auth/cmd/auth/main.go`
- `services/auth/internal/{config,secure,token,session,store,service,server}/*.go`
- `services/auth/internal/{secure,service}/*_test.go`
- `services/auth/Dockerfile`
- `services/auth/README.md`
- `.env.example` — added `AUTH_ENCRYPTION_KEY`, TLS cert/key vars, and `AUTH_SESSION_TTL`

**Key decisions made:**
- JWT signing is implemented with the Go standard library (`HS256`) instead of
  adding another dependency; Kong validates the same issuer/key shape in A-09.
- Refresh tokens are opaque random values. Only their SHA-256 digest appears in
  Redis keys; the stored session body is AES-256-GCM encrypted at rest.
- The service refuses to start without TLS cert/key paths and enforces TLS 1.3
  at the server level, matching NFR-4.
- `AUTH_SESSION_TTL` is validated to exactly `1h` so the Redis session contract
  cannot drift from DATABASE_SCHEMA.md.

**API contracts changed:** No — implements the documented auth contract.

**Known limitations:**
- Live PostgreSQL and Redis integration tests are not included yet; current tests
  cover crypto and service error mapping. Runtime DB/Redis paths are exercised
  through the same pgx/go-redis conventions as A-06/A-07.
- Kong and the auth service must share the same production JWT secret.

**Verification:** `gofmt -w .`, `go test -mod=readonly ./...`,
`go vet -mod=readonly ./...`, and `go build -mod=readonly ./...` all pass in
`services/auth`.

---

### [DONE] A-09 — Kong API Gateway Configuration
**Session:** 6
**Completed by:** Agent A
**Commit:** `session(A-08..A-11): auth gateway sonarqube`

**What was built:**
Replaced the placeholder Kong DB-less configuration with a real route table for
the public API surface. Auth endpoints route to `auth-service`; search,
wishlist, alerts, savings, and truth-score routes route to the BFF; product
history routes route to the history service. Added global rate limiting and JWT
validation on non-auth service routes.

**Files created/modified:**
- `api-gateway/kong.yml`
- `api-gateway/plugins/README.md`

**Key decisions made:**
- Auth routes are intentionally left without the JWT plugin so users can
  register, log in, and refresh.
- Regex routes are used for product subroutes so `/history` and `/truth-score`
  do not collide.
- Kong is configured with issuer/key `mergemarket-auth`, matching the auth
  service JWT `iss` claim.

**API contracts changed:** No.

**Known limitations:**
- The BFF service must implement wishlist, alerts, savings, search, and
  truth-score endpoints behind Kong.
- Production deployment must ensure Kong's JWT secret matches `JWT_SECRET`.

---

### [DONE] A-11 — SonarQube on VPS
**Session:** 6
**Completed by:** Agent A
**Commit:** `session(A-08..A-11): auth gateway sonarqube`

**What was built:**
Added a K3s SonarQube manifest and setup documentation for the VPS instance at
`http://95.111.228.35:9000`. The manifest creates a dedicated observability
namespace, persistent volume claims for SonarQube data/extensions/logs, a
single-replica deployment, health probes, and a NodePort service.

**Files created/modified:**
- `infra/k3s/sonarqube.yml`
- `docs/sonarqube-setup.md`

**Key decisions made:**
- Persistence is split across the standard SonarQube paths:
  `/opt/sonarqube/data`, `/opt/sonarqube/extensions`, and
  `/opt/sonarqube/logs`.
- The manifest uses `sonarqube:community` and leaves database hardening for the
  existing VPS/local compose setup; A-11's documented requirement was to verify
  persistence config and document token generation.
- Documentation includes deploy/reconcile commands, PVC verification, firewall
  opening for port 9000, and Jenkins CI token generation steps.

**API contracts changed:** No.

**Known limitations:**
- The remote VPS was not modified from this local session; the manifest and docs
  are committed for application on the VPS. The task note says SonarQube is
  already running at v26.5.0, so this captures the reproducible configuration.

---

### [DONE] A-12 — Grafana Dashboard on VPS
**Session:** 7
**Completed by:** Agent A
**Commit:** `session(A-12): grafana quality & pipeline dashboard`

**What was built:**
A Grafana deployment for the MergeMarket **quality & pipeline** dashboard,
sourced from the **SonarQube API** and the **GitHub Actions API**. Grafana runs
on the VPS at `http://95.111.228.35:3000` (NodePort `30300` → container `3000`),
in the shared `mergemarket-observability` namespace created by A-11.

Deliverables:
- `infra/k3s/grafana.yml` — self-contained K3s manifest: namespace, a 2Gi PVC,
  two provisioning ConfigMaps (datasources + dashboard provider), the Grafana
  Deployment (image `grafana/grafana:11.1.0`, installs the Infinity plugin via
  `GF_INSTALL_PLUGINS`, runs as uid 472, `/api/health` probes), and a NodePort
  Service.
- `infra/grafana/provisioning/datasources/datasources.yml` — SonarQube + GitHub
  Infinity datasources (mirrored into the manifest ConfigMap).
- `infra/grafana/provisioning/dashboards/provider.yml` — file-based provider.
- `infra/grafana/dashboards/mergemarket-quality.json` — 7-panel dashboard:
  coverage %, open bugs, vulnerabilities, pipeline pass/fail rate (donut),
  coverage-over-time, bugs/vulnerabilities-over-time, and build-duration-over-
  time. Two dashboard variables (`sonar_project`, `gh_repo`) retarget it
  without editing JSON.
- `.env.example` — Grafana block (`GRAFANA_PORT`, admin creds, `GITHUB_TOKEN`;
  reuses the existing `SONAR_HOST_URL`/`SONAR_TOKEN`).
- `docs/grafana-setup.md` — full runbook (secret creation, dashboard ConfigMap
  command, deploy/reconcile, firewall, local-Docker alternative, limitations).

**Verification:** dashboard JSON parses (`json.load`); all three YAML files and
the 6-document `grafana.yml` parse (`yaml.safe_load_all`), including the two
embedded ConfigMap YAML payloads. The live VPS was **not** modified from this
local session (unreachable from local dev) — artefacts are committed for
application on the VPS, matching the A-11 approach.

**Key decisions made:**
- **Infinity datasource (`yesoreyeram-infinity-datasource`)** for both APIs —
  neither SonarQube nor GitHub has a native Grafana datasource; Infinity is the
  standard $0 way to read arbitrary JSON HTTP APIs and satisfies the task's
  "SonarQube API and GitHub API data sources" literally.
- **No secrets committed.** Admin password + SonarQube/GitHub tokens come from a
  `grafana-secrets` K8s Secret the operator creates on the VPS (documented).
  `.env.example` holds only placeholders.
- **Dashboard JSON lives in `infra/grafana/dashboards/`** (per the task output)
  and is loaded via a `grafana-dashboards` ConfigMap generated from that folder
  with one documented `kubectl create configmap --from-file` command — keeping
  the large JSON out of the manifest YAML.
- **SonarQube token auth** is sent as the HTTP Basic username (empty password),
  per SonarQube's documented scheme; **GitHub** uses a Bearer fine-grained PAT.

**API contracts changed:** No.

**Known limitations:**
- **Pipeline panels (pass/fail rate, build duration) depend on GitHub Actions
  run data.** CI was migrated to Jenkins (`Jenkinsfile`, A-10) and the Actions
  workflow dir was deleted previously, so those two panels are empty until an
  Actions workflow runs again (or the panels are re-pointed at a Jenkins
  source). This is the same CI-source reconciliation flagged in A-04/A-10.
- Dashboard defaults `sonar_project=mergemarket`; confirm against the final
  `sonar.projectKey` (no `sonar-project.properties` exists in the repo yet).
- Provisioning YAML is duplicated between `infra/grafana/provisioning/` and the
  ConfigMaps in `grafana.yml` (the manifest is intentionally self-contained);
  edits must be mirrored.

---

## Entry Format

When you complete a task, add an entry using this format:

```markdown
---

### [DONE] A-XX — Task Name
**Session:** X
**Completed by:** Agent A
**Commit:** `session(A-XX): task name`

**What was built:**
Brief description of what was created or configured.

**Files created/modified:**
- `path/to/file1`
- `path/to/file2`

**Key decisions made:**
Any architectural or implementation decisions and the reason for them.

**API contracts changed:**
Yes / No — if Yes, describe what changed so Agent B can update mocks.

**Known limitations:**
Anything left incomplete, deferred, or that needs follow-up.
```
