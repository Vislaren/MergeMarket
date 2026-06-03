# Proxy Validator Service (Go)

Continuously scrapes public proxy lists, tests each proxy against a real
endpoint, and writes working IPs to Redis (`proxy_pool`, TTL 5min).
Implements the politeness protocol (adaptive random delays).

- **Port:** 8086
- **Endpoints:** `GET /health`
- **Task:** A-04

_Implementation pending._
