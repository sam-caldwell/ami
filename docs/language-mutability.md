# Data Mutability (Imperative Subset 6.4)

AMI defaults to immutable code regions. Assignments are only allowed within explicit `mut { ... }` sections.

Rules

- Immutable by default: any `=` assignment outside a `mut { ... }` block emits `E_MUT_ASSIGN_OUTSIDE`.
- Nested blocks: nested `{ ... }` inside a `mut { ... }` remain mutable until the matching `}` closes the mut region.
- After block: assignments after the closing brace of a mut block are immutable again and will emit `E_MUT_ASSIGN_OUTSIDE`.

Notes

- The parser captures function body tokens, and the semantic analyzer tracks `{`/`}` depth with a separate counter for mut-initiated blocks, allowing nested blocks within a mut region.
- This scaffold enforces region-based mutability; finer-grained variable-level controls can be layered later if/when the language adds them.

Tests

- See `src/ami/compiler/sem`:
  - `TestMutability_AssignOutsideMut_Error`
  - `TestMutability_AssignInsideMut_OK`
  - `TestMutability_NestedBlocksInsideMut_OK`
  - `TestMutability_AfterMutBlock_Error`

