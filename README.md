# MergeMarket

Real-time e-commerce price aggregation platform built on a $0/month
bootstrap budget. A user searches once and receives live prices from
50–150 stores within 5 seconds.

## Architecture

Microservices in Go behind a Kong API gateway, with a Flutter mobile
client. PostgreSQL + TimescaleDB for storage, Redis for cache/sessions,
K3s on a self-hosted VPS. Full details in
[`project_docs/architecture/ARCHITECTURE.md`](project_docs/architecture/ARCHITECTURE.md).

## Repository Layout

```
api-gateway/        Kong declarative config and custom plugins
apps/mobile/        Flutter client (Agent B)
services/           Go microservices (Agent A)
  auth/             JWT auth + sessions            :8081
  bff/              Backend-for-Frontend           :8082
  scraper-service/  Config-driven scraper engine   :8083
  normalization/    Raw → standard product schema  :8084
  history/          Price snapshots + alerts       :8085
  proxy-validator/  Proxy health checking          :8086
  mock-server/      Local dev mocks (not deployed)
infra/              K3s manifests + Terraform IaC
shared/             Shared Go packages
docs/               Setup + testing artefacts
project_docs/       Architecture, API contracts, DB schema, UI specs
.agents/            Agent A / Agent B tracking files
```

## Documentation

- [Architecture](project_docs/architecture/ARCHITECTURE.md)
- [API Contracts](project_docs/api/API_CONTRACTS.md)
- [Database Schema](project_docs/database/DATABASE_SCHEMA.md)
- [VPS Setup](docs/vps-setup.md)

## Local Development

Spin up the local stack (PostgreSQL/TimescaleDB, Redis, Kong, SonarQube)
with `docker-compose up` once `docker-compose.yml` lands (task A-02).
Copy `.env.example` to `.env` and fill in the values first.
