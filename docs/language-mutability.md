# Data Mutability (Imperative Subset 6.4)

AMI does not support Rust-like `mut { ... }` blocks. Mutability is explicit and local:
- Use `*` on the left-hand side to mark a mutating assignment, e.g., `*x = value`.
- Wrap side-effectful expressions in `mutate(expr)` to signal mutation where appropriate.

Rules

- Unmarked assignment: any `x = y` without `*` on the left-hand side emits `E_MUT_ASSIGN_UNMARKED`.
- `*` is a marker, not deref: unary `*` is not a pointer dereference. It is only permitted on the assignment left-hand side; using `*` in an expression position emits `E_STAR_MISUSED`.

Notes

- The parser builds a simple statement AST; the semantic analyzer verifies that assignments are marked with `*` and flags any legacy `mut { ... }` usage as unsupported.
- This scaffold enforces region-based mutability; finer-grained variable-level controls can be layered later if/when the language adds them.

Tests

- See `src/ami/compiler/sem`:
  - `TestMutability_AssignOutsideMut_Error`
  - `TestMutability_AssignInsideMut_OK`
  - `TestMutability_NestedBlocksInsideMut_OK`
  - `TestMutability_AfterMutBlock_Error`
