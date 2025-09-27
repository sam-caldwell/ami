# E2E CLI Tests

This suite exercises the built `ami` binary via stdin/stdout/stderr to validate end‑to‑end behavior of `ami mod *` subcommands. Tests are hermetic and stage inputs under `build/test/e2e/...`.

Scenarios
- mod clean: removes and recreates the package cache; JSON/human behaviors and error cases.
- mod list: lists cache entries (packages and versions) in JSON and human modes.
- mod get: fetches local workspace packages into the cache and updates `ami.sum`; sad path checks for invalid paths.
- mod sum: validates `ami.sum` against cache and can fetch from `file+git://` sources; sad path for missing `ami.sum`.
- mod update: copies workspace packages to cache and refreshes `ami.sum`; sad path for missing workspace.
- mod audit: audits workspace requirements vs `ami.sum` and cache in JSON/human modes (see ami_mod_audit_test.go).

How to run
- Build and run E2E tests:
  - `make e2e-test`
- Or from the repo root:
  - `go build -o build/ami ./src/cmd/ami`
  - `go test -v ./tests/e2e`

Notes
- Tests set `AMI_PACKAGE_CACHE` to a subdirectory of the test workspace to avoid global side effects.
- `stdin` is always attached to an empty reader to ensure non‑interactive operation.
- Outputs are asserted on stable JSON fields or concise human summary lines.
