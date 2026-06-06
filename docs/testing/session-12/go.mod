// Module for the Session-12 test suites (unit + integration) covering Agent A's
// A-12 — Grafana Dashboard on VPS.
//
// Self-contained so the suites run independently of the service modules. The
// unit suite statically validates the committed A-12 artefacts (K8s manifest,
// provisioning YAML, dashboard JSON); the integration suite (build tag
// `integration`) smoke-tests a live Grafana once a cluster/endpoint is reachable.
module github.com/Vislaren/MergeMarket/docs/testing/session-12

go 1.22

require (
	github.com/stretchr/testify v1.9.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)
