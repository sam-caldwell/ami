# Memory Model (Ch. 2.4)

This repository implements the Memory Model feature at a scaffold level to enable
compile-time safety and IR visibility:

- Allocation domains: `event`, `state`, `ephemeral`.
- Ownership semantics: `Owned<T>` parameters must be released or transferred.
- Per-VM runtime scaffold: counters per domain with RAII-style Handles.

Highlights

- IR (`ir.v1`) includes optional parameter annotations with `ownership` and `domain`.
- Lints enforce:
  - Owned<T> must be released/transferred; no double-release or use-after-release.
  - Cross-domain references into state from non-state domains are forbidden.
- Tests cover owned transfer, lifetime errors, and cross-domain violations.

Notes

- The runtime memory manager is a scaffold for accounting and deterministic
  release in tests; it is not a production allocator.
- Domain analysis operates at token level and focuses on key patterns such as
  `*st = &ev`. Future work can deepen flow analysis and aliasing detection.

