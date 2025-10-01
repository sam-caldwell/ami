# Makefile Targets

Build and test helpers.

Key targets
- `make build`: build CLI to `build/ami`.
- `make test`: verbose `go test ./...` with race and coverage.
- `make coverage-short`: quick coverage on key packages.
- `make gen-diag-codes`: regenerate `docs/diag-codes.md` from code.
- `make examples`: build examples under `examples/` and stage outputs under `build/examples/`.

Notes
- CI uses `go vet ./...` and `go test ./...`; see `docs/toolchain/ci.md`.
