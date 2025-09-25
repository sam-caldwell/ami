# Set Type (Imperative Subset 6.4)

This document describes the `set<T>` container type constraints in the AMI language as implemented in the compiler semantics.

Constraints

- Arity: `set<T>` must have exactly one type argument. Using zero or more than one emits `E_SET_ARITY`.
- Element type restrictions (comparable requirement):
  - Not a pointer type (no `*T`).
  - Not a slice type (no `[]T`).
  - Not another container type: `map`, `set`, or `slice`.
  - Not a generic instantiation (element type must not itself have generic arguments).

Examples

- Valid: `set<int>`, `set<string>`, `set<MyEnum>`, `set<MyStruct>`
- Invalid:
  - `set<>` or `set<A,B>` → `E_SET_ARITY`
  - `set<*T>` → `E_SET_ELEM_TYPE_INVALID`
  - `set<[]byte>` → `E_SET_ELEM_TYPE_INVALID`
  - `set<map<string,int>>` → `E_SET_ELEM_TYPE_INVALID`
  - `set<Other<X>>` → `E_SET_ELEM_TYPE_INVALID`

Testing

- Happy and sad path unit tests are under `src/ami/compiler/sem/set_semantics_test.go` and cover arity and element-type constraints.

