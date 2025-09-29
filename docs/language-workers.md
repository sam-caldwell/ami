# AMI Worker Signatures

This document defines canonical worker function signatures for pipeline `Transform` steps, per the AMI docx. It replaces
legacy suggestions that used an explicit `State` parameter; AMI workers do not accept raw state handles. Ambient node
state resides behind runtime-managed boundaries and capabilities.

Canonical Signature

- Worker: `func(Event<T>) (Event<U>, error)`
  - Input: a strongly typed `Event<T>` payload.
  - Output: a strongly typed `Event<U>` payload and an `error`.
  - No raw pointers or address-of; memory safety rules apply (see 2.3.2). Event payload types must be pointer‑free;
    see
    language-events.md.
  - Factories: functions named `New*` are treated as factories and are not subject to worker signature checks.

Constraints & Semantics

- No Decorators: workers must not carry decorators; these are reserved for declarations and tooling hints. The compiler
  emits `E_DECORATOR_ON_WORKER` otherwise.
- No Raw Pointers: AMI forbids `&` and pointer dereference. Mutating assignments require `*` on LHS and `mutate(expr)`
  for side‑effects, as specified in memory safety rules.
- Capability Boundaries: I/O and trust policies are enforced per node, not via ambient parameters.
- Type Inference: `T`/`U` may be inferred locally from usage; diagnostics ensure mismatches are reported
  deterministically.

Examples

```
// Good: simple identity transform
default package app
func Identity(ev Event<string>) (Event<string>, error) {
  return ev, nil
}

// Good: factory (New* exempt from worker signature checks)
func NewHasher() (Hasher, error) { /* ... */ }

// Bad: wrong results
func Bad(ev Event<int>) (int, error) { /* ... */ }
// E_WORKER_SIGNATURE: want func(Event<T>)->(Event<U>, error)
```

Diagnostics

- `E_WORKER_UNDEFINED`: referenced worker is not declared locally (simple scope).
- `E_WORKER_SIGNATURE`: worker does not match `func(Event<T>)->(Event<U>, error)`.
- `E_DECORATOR_ON_WORKER`: workers cannot be decorated.

Notes

- Node state and runtime context are not passed as parameters; use compiler/runtime facilities and capabilities instead.
- The compiler enforces these constraints during semantics analysis; see `sem.AnalyzeWorkers`.
