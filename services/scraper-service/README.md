# Scraper Service (Go)

Config-driven worker-queue scraper engine. Each store is defined by a
`StoreConfig` JSON/YAML file in `configs/`. Workers consume jobs, load the
relevant config, execute the scrape, and write raw results to the
normalisation queue. Implements a Circuit Breaker that trips after
consecutive 403/429 failures.

- **Port:** 8083
- **Endpoints:** `GET /health`
- **Task:** A-05
- **Configs:** `configs/` — one `StoreConfig` file per store

_Implementation pending._
