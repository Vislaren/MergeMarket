## Oracle — Integration — Session 11

| Input | Condition | Expected Output | Source of Truth |
|-------|-----------|----------------|----------------|
| Login with valid creds (live) | mock/real Auth up | token persisted; Home shows Log out | USER_FLOWS Flow 1 |
| Register fresh email (live) | mock/real Auth up | 201; authenticated Home | API_CONTRACTS Auth |
| Tap price-drop push (live) | app backgrounded | opens `/product/{id}` | USER_FLOWS Flow 6 |
| `GET /detail` (live BFF + upstream) | upstream up | merged history+truth+offers, sorted | B-09 spec + B-04 |
| `GET /alerts` via BFF (live) | no BFF handler | upstream body returned unchanged | B-09 spec (forwarding) |
| any backend unreachable | transport failure | `MMErrorState` / `ApiException(network)` | API error shape |
