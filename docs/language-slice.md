# Slice Type (Imperative Subset 6.4)

This document describes the `slice<T>` container type and `[]T` shorthand in the AMI language as implemented in the compiler semantics.

Forms

- Generic: `slice<T>`
- Bracket: `[]T`

Constraints

- Arity (generic form): `slice<T>` must have exactly one type argument. Using zero or more than one emits `E_SLICE_ARITY`.
- Element type: No special restrictions beyond nested type validation (e.g., `map` key rules apply when present). Pointer types are not part of AMI; generic element types are allowed, e.g., `[]Event<U>`.

Examples

- Valid: `[]int`, `[]Event<string>`, `slice<byte>`, `[]map<string,int>` (map key rules still apply)
- Invalid: `slice<>`, `slice<A,B>` → `E_SLICE_ARITY`

Testing

- Unit tests for arity and bracket/generic forms live under `src/ami/compiler/sem/slice_semantics_test.go`.
