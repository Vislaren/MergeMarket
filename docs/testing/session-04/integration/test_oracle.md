## Oracle — A-04 Proxy-Validator Service (integration / live binary)

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|-----------------|-----------------|
| `GET /health` | binary running, Redis + source DOWN | `200`, `{status:"ok", service:"proxy-validator", version}` | API_CONTRACTS.md + NFR-2 (health independent of loop) |
| Fake source returns proxy `H:P`; fake proxy answers `204` | one cycle completes | `/stats` → `has_run:true`, `working >= 1` | A-04 validate path |
| `SMEMBERS proxy_pool` | after a successful cycle | contains `H:P` | DATABASE_SCHEMA §3 (Redis Set of `ip:port`) |
| `TTL proxy_pool` | after a successful cycle | `0 < ttl <= 5m` | DATABASE_SCHEMA §3 (5-minute TTL) |
| Redis unreachable at startup | binary launches | warns and keeps serving `/health`; retries next cycle | A-04 resilience / NFR-2 |
| Unreachable proxy source | cycle runs | logs "all sources failed", no crash; `/health` stays up | NFR-2 |

**Note on the fake proxy:** for plain-HTTP targets Go's proxy transport sends the
absolute-URI request to the proxy, so a proxy that replies `204` to anything is a
faithful stand-in for a "working" proxy without contacting any real upstream.
