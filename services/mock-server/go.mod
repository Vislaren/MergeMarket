// Mock Server (B-02) — local-dev-only HTTP server that serves the canonical
// API contracts (project_docs/api/API_CONTRACTS.md) with static fixtures so the
// Flutter app and BFF can be developed offline before Agent A's real services
// exist. Intentionally dependency-free (stdlib only) so it always builds and
// runs without a network round-trip.
module github.com/Vislaren/MergeMarket/services/mock-server

go 1.22
