# Gaps Reconciliation

This document captures known gaps or clarifications between the authoritative docx and the YAML tracker, and proposes next actions. The intent is to make gaps explicit and track them to closure.

Scope: As of this update, the repository’s tracker (work_tracker/specification-v.0.0.1.yaml) reflects implemented features. The items below are either in‑progress cross‑cutting gates or candidate enhancements that improve clarity and developer UX.

Known items

- Cross‑cutting (status):
  - CC‑1 Coverage Gate — COMPLETE: ≥0.80 coverage on changed packages enforced via CI (`.github/workflows/ci.yml`) using `scripts/coverage_gate.sh`.
  - CC‑2 No raw pointers in public ABI — COMPLETE: enforced via backend safety checks (LLVM emitter + tests). Backends reject pointer params/results at function boundaries and map handle‑like types to i64 in public signatures. Driver surfaces violations as E_LLVM_EMIT. New backends must gate by the same rule.
  - CC‑3 RAII Compliance — COMPLETE: SEM analyzer detects leaks, double‑release, release/assign/use after transfer; driver lowers Owned with zeroize on release. Tests cover leak/transfer/double‑release and runtime zeroization.

- Generics diagnostics (candidate enhancements):
  - Add E_GENERIC_ARITY_MISMATCH for wrong number of generic type arguments (e.g., Owned<>) to improve feedback beyond generic type mismatch.
  - Enrich diagnostics data payloads with {expected, actual, paramName} where available for call/return mismatches.

- Sandbox configuration (CLI ergonomics):
  - exec.SandboxPolicy exists for simulated runtime tests. Consider adding `ami run` flags to allow specifying fs/net/device capability policies for examples and demos.

- Examples/docs cohesion:
  - Minimal POP example validated in tests (multi‑target). `examples/simple` and `examples/complex` are available; docs/toolchain/examples.md describes usage. Optionally add a short POP quickstart snippet referencing `examples/simple`.

- Diagnostics reference upkeep:
  - `docs/diag-codes.md` can be regenerated with `make gen-diag-codes`. Keep in sync after introducing new diagnostic codes.

AMI stdlib status and plan
- signal/time: Go prototypes exist under `src/ami/stdlib/{signal,time}` but are not AMI stdlib implementations. The AMI stdlib must be provided as language-level packages or intrinsics available to `.ami` code.
  - Action: Revert S-4 (signal) and S-5 (time) to `ready` in the tracker.
  - Plan:
    - Define AMI package surfaces (signatures) for `signal` and `time` in `.ami` terms.
    - Add compiler hooks or intrinsic lowering so that `signal.Register`, `time.Sleep`, `time.Now`, etc. are callable from `.ami` and lower to runtime/platform primitives.
    - Provide `.ami` stubs and tests that import and use these packages, validating parser/semantics → IR lowering paths.

Status

- The YAML tracker reflects current features and ready/completed statuses. This file will evolve as new diagnostics or CLI ergonomics are added.
