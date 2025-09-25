# Pointers and Addresses (Imperative Subset 6.4 / 2.3.2)

This document describes pointer and address semantics as implemented in the compiler’s semantic checks.

Operators

- Address-of: `&x`
- Dereference: `*x`

Rules

- Safe deref: Using `*x` is only allowed inside a block guarded by a nil-check for `x`, e.g., `if x != nil { *x }` or `if nil != x { *x }`. Otherwise emits `E_DEREF_UNSAFE`.
- Invalid deref operand: `*` cannot be applied to literals or `nil`. Emits `E_DEREF_OPERAND`.
- Address-of restrictions: `&` cannot be applied to literals or `nil`. Emits `E_ADDR_OF_LITERAL`.
- Assignment mutability: Assignments still require `mut { ... }` blocks; pointer writes via `*x = ...` outside `mut` will also emit `E_MUT_ASSIGN_OUTSIDE` (existing rule).

Notes

- The analysis is token-based (scaffold). A nil-guard is recognized when a block `{ ... }` is immediately preceded by `x != nil` or `nil != x`.

Tests

- See `src/ami/compiler/sem/pointer_semantics_test.go` for happy/sad cases.

