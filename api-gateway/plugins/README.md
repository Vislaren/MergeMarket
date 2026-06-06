# Kong Plugin Configuration

MergeMarket uses Kong Community Edition in DB-less mode. Built-in plugins are
declared directly in `api-gateway/kong.yml`.

Current plugins:

- `rate-limiting` globally limits client traffic.
- `jwt` protects all non-auth public API routes.

No custom Lua plugins are required for A-09.
