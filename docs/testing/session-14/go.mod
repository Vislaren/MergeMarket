// QE test module for Session 14 — independent verification of Agent A's
// session-8 services: A-14 (search) and A-16..A-18 (userdata). Kept as its own
// module (matching session-04/12) because Go forbids importing a service's
// internal/ packages from outside its module, so the suite verifies via the
// toolchain (build/test/vet/gofmt as subprocesses), structural source assertions
// of the contract invariants, and — for integration — by building and running
// the real service binaries.
module github.com/Vislaren/MergeMarket/docs/testing/session-14

go 1.22

require github.com/stretchr/testify v1.9.0

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
