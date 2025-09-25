# Imperative Type Checking (Imperative Subset 6.4)

This scaffold implements minimal type checks for imperative constructs using parameter types:

Checks

- Assignment type match: For simple forms `x = y` where `x`, `y` are function parameters or literals, the checker ensures left and right types match (emits `E_ASSIGN_TYPE_MISMATCH` when they do not).
- Address-of is disallowed: `&` anywhere in source emits `E_PTR_UNSUPPORTED_SYNTAX`.

Scope

- The type environment consists of function parameters (name → type). Local variable declarations are not parsed in this scaffold; expressions are recognized from tokens in simple forms.

Tests

- See `src/ami/compiler/sem/imperative_typecheck_test.go` for happy/sad cases.
