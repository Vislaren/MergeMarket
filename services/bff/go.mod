// BFF — Backend-for-Frontend (B-09). Sits between Kong and the Flutter client:
// it forwards most read endpoints straight through to the upstream services and
// shapes one aggregate view (product detail = history + truth-score + offers)
// so the mobile app makes a single call. No business logic lives here.
//
// Intentionally dependency-free (Go standard library only), matching the
// mock-server (B-02): it always builds and its tests run offline against an
// httptest upstream.
module github.com/Vislaren/MergeMarket/services/bff

go 1.22
