# MergeMarket — Architecture

## 1. System Overview

MergeMarket is a microservices-based, real-time price aggregation platform.
A user searches for a product once and receives live prices from 50–150
stores within 5 seconds. The system runs on a $0/month bootstrap budget
using a self-hosted VPS, free proxies, and open-source tooling.

```
Flutter App
    │
    ▼
Kong API Gateway ──→ Auth Service
    │
    ├──→ BFF Service
    │       │
    │       ├──→ Scraper Service ──→ Proxy Validator ──→ Free Proxies
    │       │         │
    │       │         ▼
    │       │   Normalization Service ──→ Affiliate Link Injection
    │       │         │
    │       │         ▼
    │       │   PostgreSQL / TimescaleDB / Redis
    │       │
    │       └──→ History Service ──→ TimescaleDB
    │
    └──→ Notification Worker ──→ Firebase / APNs
```

---

## 2. Service Breakdown

### API Gateway — Kong Community Edition
- Single entry point for all client requests
- Handles rate limiting, JWT validation, and routing
- Config: `api-gateway/kong.yml`

### Auth Service — Go
- Issues and validates JWT tokens
- Manages user registration, login, and refresh
- Sessions cached in Redis (TTL: 1 hour)
- Data encrypted at rest (AES-256), in transit (TLS 1.3)
- Port: `8081`

### BFF (Backend-for-Frontend) — Go
- Shapes and aggregates data specifically for the Flutter client
- No business logic — only formatting and forwarding
- Port: `8082`

### Scraper Service — Go
- Config-Driven Worker-Queue pattern
- Each store defined by a `StoreConfig` JSON/YAML file
- Workers consume jobs, load config, execute scrape
- Circuit Breaker: trips after consecutive 403/429 failures
- Targets internal JSON APIs where available
- Port: `8083`

### Normalization Service — Go
- Converts raw scrape output to standard schema
- Schema: `{ price, currency, title, image_url, shipping, affiliate_url }`
- Includes Affiliate Link Injection module (FR-6)
- Port: `8084`

### History Service — Go
- Records daily price snapshots to TimescaleDB
- Runs scheduled heartbeat scrapes on followed products
- Triggers price-drop alert when threshold is crossed
- Port: `8085`

### Proxy Validator Service — Go
- Scrapes public proxy lists continuously
- Tests each proxy against a real endpoint
- Writes working IPs to Redis (TTL: 5 minutes)
- Implements politeness protocol (adaptive random delays)
- Port: `8086`

---

## 3. Infrastructure

| Component | Tool | Location |
|---|---|---|
| Hosting | Existing VPS + K3s | Contabo VPS |
| Container runtime | Docker + Docker Compose | VPS + local |
| Orchestration | K3s (lightweight Kubernetes) | VPS |
| API Gateway | Kong Community Edition | VPS |
| Firewall / WAF | CrowdSec + NGINX | VPS |
| CI/CD | GitHub Actions | Cloud |
| Code quality | SonarQube Community | VPS port 9000 |
| Monitoring | Grafana | VPS port 3000 |
| IaC | Terraform | `infra/terraform/` |

---

## 4. Data Architecture

### PostgreSQL — Relational Data
Stores users, store configs, products, wishlist items, alerts, and return policies.

### TimescaleDB — Time-Series Data
Stores price history as a hypertable partitioned by time.
Enables efficient queries like "price over last 6 months".

### Redis — Cache & Sessions
Five cache layers: Store Metadata (24hr), Proxy Health (5min),
Auth Sessions (1hr), Product Scores (24-48hr), Search Results (5min-24hr).

---

## 5. Repository Folder Structure

```
mergemarket/
├── api-gateway/                  # Kong configuration and plugins
│   ├── kong.yml                  # Declarative Kong config (routes, plugins)
│   └── plugins/                  # Custom Kong plugin configs
├── apps/
│   └── mobile/                   # Flutter project
│       ├── lib/
│       │   ├── screens/          # One file per screen
│       │   ├── widgets/          # Reusable components (MM* components)
│       │   ├── providers/        # Riverpod state providers
│       │   ├── services/         # API calls and external integrations
│       │   ├── models/           # Data classes / DTOs
│       │   └── theme/            # Colours, typography, spacing tokens
│       ├── test/
│       │   ├── unit/             # Unit tests per feature
│       │   └── mocks/            # Mock data files for offline dev
│       └── integration_test/     # Flutter integration tests
├── services/
│   ├── scraper-service/          # Worker-Queue engine (Go)
│   │   └── configs/              # StoreConfig JSON/YAML (one per store)
│   ├── normalization/            # Data cleaning and standardization (Go)
│   ├── bff/                      # Backend-for-Frontend service (Go)
│   ├── auth/                     # JWT and user session management (Go)
│   ├── history/                  # Price history and analytics (Go)
│   ├── proxy-validator/          # Proxy health checking and rotation (Go)
│   └── mock-server/              # Local dev mock server (Go) — not deployed
├── infra/
│   ├── k3s/                      # Kubernetes manifests for all services
│   └── terraform/                # Infrastructure-as-Code files
├── shared/                       # Shared Go packages (models, helpers)
├── docs/
│   ├── testing/                  # Agent B test artefacts (session-XX/)
│   └── vps-setup.md              # VPS setup documentation
├── .agents/                      # Agent instruction and tracking files
│   ├── Agent_A/
│   ├── Agent_B/
│   └── project-docs/
├── .github/
│   └── workflows/
│       └── ci.yml                # GitHub Actions pipeline
├── .env.example                  # All required environment variables
├── docker-compose.yml            # Local dev environment
└── sonar-project.properties      # SonarQube project config
```

---

## 7. Image Handling

Images are **never stored on the VPS**. Only the URL string is persisted.
The Flutter app fetches images directly from each retailer's CDN,
eliminating VPS storage and bandwidth costs entirely.

---

## 8. Key Non-Functional Requirements

| ID | Metric | Requirement |
|---|---|---|
| NFR-1 | Latency | Results within 5 seconds for 80% of queries |
| NFR-2 | Resilience | Single scraper failure must not halt others |
| NFR-3 | Accuracy | 98% price match rate with source at scrape time |
| NFR-4 | Security | AES-256 at rest, TLS 1.3 in transit |
| NFR-5 | Scalability | Handle 5,000 concurrent search requests |

---

## 9. Caching Strategy (Stale-While-Revalidate)

Cached data is served immediately to the client while a background worker
refreshes it. This guarantees sub-5s latency even during live scrape cycles.

---

## 10. Cost Growth Path

Phase 1 — $0: Free proxies, self-hosted everything
Phase 2 — Revenue reinvested: Upgrade to paid datacenter proxies
Phase 3 — Scale: Managed PostgreSQL (Supabase/Neon), scale K3s cluster
