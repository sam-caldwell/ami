# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to semantic versioning when releases are tagged.

## Unreleased

### Added
- CI coverage gate (CC‑1): enforce ≥0.80 coverage on changed packages.
  - Script: `scripts/coverage_gate.sh` with env tuning:
    - `THRESHOLD` (global), `PKG_THRESHOLDS` (regex=threshold overrides),
    - `SKIP_PKGS` (regex to skip), `EXTRA_EXCLUDES` (additional diff excludes).
  - Workflow: `.github/workflows/ci.yml` runs `go vet`, `go test`, and the coverage gate.
- Diagnostics: alias‑qualified call `expectedPos` tests for cross‑package calls.
- Docs: `docs/diag-codes.md` notes Optional/Union support for arity `path/fieldPath` traversal.

### Changed
- Cross‑cutting statuses: Marked CC‑1, CC‑2, CC‑3 as COMPLETE in the tracker and gaps doc.
  - `work_tracker/specification-v.0.0.1.yaml`: CC‑1/CC‑2/CC‑3 → complete.
  - `docs/gaps.md`: updated statuses and descriptions.

### Fixed
- Backend/driver enforcement for CC‑2: Verified tests for E_LLVM_EMIT on pointer‑like ABI exposure remain green.

### Notes
- The coverage gate ignores common generated/testdata paths (testdata, generated, mocks, zz_generated, *_mock.go) by default and supports custom excludes via `EXTRA_EXCLUDES`.

