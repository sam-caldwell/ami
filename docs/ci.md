# CI and Coverage Gate

This repository includes a CI workflow that enforces a per-package coverage gate on changed code paths.

Overview
- CI workflow: `.github/workflows/ci.yml`
- Coverage script: `scripts/coverage_gate.sh`
- Behavior: For packages under `src/` that have changed since the base ref, run tests with coverage and require coverage ≥ threshold. If a changed package has no `*_test.go`, the gate fails.

Environment variables
- `THRESHOLD`: Global minimum coverage for changed packages (default: `0.80`).
- `SKIP_PKGS`: Regex; skip packages whose import path matches this pattern. Example: `'^github.com/.*/visualize/ascii$'`.
- `PKG_THRESHOLDS`: Comma-separated `regex=threshold` overrides. The first matching rule wins.
  - Example: `'^github.com/.*/compiler/parser=0.90,.*codegen/llvm.*=0.75'`.
- `EXTRA_EXCLUDES`: Space/comma-separated path patterns to exclude from changed file detection.
  - Default excludes include `testdata`, `generated`, `mocks`, `*.pb.go`, `zz_generated*.go`, `*_mock.go`, and top-level `tools/`, `examples/`, `tests/`.

Examples
- Tighten parser and relax codegen/llvm locally:
  ```bash
  THRESHOLD=0.80 \
  PKG_THRESHOLDS='^github.com/sam-caldwell/ami/src/ami/compiler/parser=0.90,^github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm=0.75' \
  bash scripts/coverage_gate.sh
  ```
- Skip visualize/ascii and any experimental packages:
  ```bash
  SKIP_PKGS='^github.com/sam-caldwell/ami/src/ami/visualize/ascii$|.*/experimental/.*' \
  bash scripts/coverage_gate.sh
  ```
- Exclude additional generated paths from diff scanning:
  ```bash
  EXTRA_EXCLUDES='src/**/autogen/** src/**/third_party/**' \
  bash scripts/coverage_gate.sh
  ```

Notes
- The gate only evaluates packages that changed in the diff range, minimizing CI noise.
- Use repository/organization environment overrides to adjust thresholds for specific packages if needed.

Defaults in CI
- Global threshold: `THRESHOLD=0.80`
- Skipped packages: `SKIP_PKGS='^github.com/sam-caldwell/ami/src/ami/visualize/ascii$|.*/experimental/.*'`
- Package thresholds: `PKG_THRESHOLDS='^github.com/sam-caldwell/ami/src/ami/compiler/parser=0.90,^github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm=0.75'`
