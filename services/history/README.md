# History Service (Go)

Records daily price snapshots to TimescaleDB. Runs a scheduled heartbeat
that scrapes followed product URLs and triggers price-drop alert logic when
a threshold is crossed.

- **Port:** 8085
- **Endpoints:** `GET /api/v1/products/{id}/history`, `GET /health`
- **Task:** A-07

_Implementation pending._
