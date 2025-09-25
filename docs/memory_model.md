# AMI Memory Model (Ch. 2.4) – Scaffold

This document summarizes the Phase 2 scaffold for the AMI memory model implemented in this repository.

Domains

- Event heap: per‑event allocations during pipeline execution.
- Node‑state: long‑lived node state owned by the runtime.
- Ephemeral stack: call‑local, short‑lived allocations.

Runtime Manager

- `src/ami/runtime/memory/manager.go` provides a lightweight per‑VM manager that tracks allocations by domain and returns RAII‑like `Handle`s for deterministic release.
- The Phase 2 runtime tester (`src/ami/runtime/tester/runner.go`) composes a manager and accounts ephemeral and event allocations per case, releasing them when execution completes.

IR Annotations

- The IR schema (`src/schemas/ir_v1.go`) includes per‑parameter fields `ownership` and `domain`.
- IR lowering populates these from function signatures: `Event<T>` → `domain=event`; `State` → `domain=state`; otherwise `domain=ephemeral`. The repository uses `Owned<…>` as an internal analysis marker to model ownership (`ownership=owned`); this is not a source‑level AMI type.

Lints

- Ownership/RAII checks ensure `Owned<T>` parameters are released or transferred, catch double‑release and use‑after‑release.
- Cross‑domain reference checks (scaffold) reject assigning address‑of expressions into state; raw address‑of is rejected by the parser per AMI 2.3.2.

Notes

- AMI 2.3.2 forbids raw pointers: `&` is not allowed, and unary `*` is only a mutation marker on the left‑hand side of assignments.
- The manager is an accounting scaffold; it is not a general‑purpose allocator.
