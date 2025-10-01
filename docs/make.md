Make Targets ============

This repository includes a convenient `Makefile` for common developer tasks. Below is a concise reference of the
available targets and how to use them.

Core Targets
- `make build`: build the `ami` CLI binary to `build/ami`.
- `make clean`: remove and recreate the `build/` directory.
- `make lint`: run `go vet` across all packages.
- `make test`: run all tests `go test -v ./...`.
- `make zip`: create a repository snapshot zip `build/repo.zip` (tracked files only).
- `make bench`: run CLI microbenchmarks for `ami` subcommands.
  - Variables:
    - `BENCH`: benchmark selector (default `BenchmarkAMI_Subcommands`).
    - `BENCHTIME`: benchtime value for `go test` (default `1x`).
  - Example: `make bench BENCH=BenchmarkAMI_Subcommands BENCHTIME=3x`.

E2E Targets
- `make e2e-build`: build `build/ami` for end‑to‑end tests.
- `make e2e-test`: run all E2E tests in `tests/e2e`.
- `make e2e-one NAME=Pattern`: run subset of E2E tests matching `Pattern`.
- `make e2e-mod-audit`: run only the E2E test(s) for `ami mod audit`.
- `make e2e-mod-clean`: run only the E2E test(s) for `ami mod clean`.
- `make e2e-mod-list`: run only the E2E test(s) for `ami mod list`.
- `make e2e-mod-get`: run only the E2E test(s) for `ami mod get`.
- `make e2e-mod-sum`: run only the E2E test(s) for `ami mod sum`.
- `make e2e-mod-update`: run only the E2E test(s) for `ami mod update`.

Utilities
- `make test-hotspots`: list packages with no tests and `.go` files missing paired `_test.go` files.
- `make examples`: build all example workspaces under `examples/` and stage outputs in `build/examples/`.

Notes
- All targets operate within the repository root. The CLI is built from `./src/cmd/ami` and outputs to `./build/ami`.
- Benchmarks run isolated sandboxes for CLI subcommands; see `src/cmd/ami/bench_subcommands_test.go` for details.
