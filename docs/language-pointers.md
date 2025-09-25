# Memory Safety and “*” (AMI 2.3.2)

AMI does not expose raw pointers or process addresses. The `&` operator (address-of) is not part of the language, and unary `*` is not a dereference.

Usage of `*`:

- When `*` appears on the left-hand side of an assignment, it explicitly marks a mutating assignment in contexts where the compiler requires clarity. This does not imply pointer dereferencing.
- Outside this left-hand marking, `*` has no pointer semantics in AMI.

Compiler behavior:

- `&` anywhere in source is rejected (`E_PTR_UNSUPPORTED_SYNTAX`).
- Unary `*` is not a dereference; using it outside the assignment left-hand side is not supported.
- Pointer type markers in signatures are not part of AMI (e.g., `*State` is invalid). Use `State` as a capability argument.
  Mutability remains explicit by language rules (see `docs/language-mutability.md`).

Notes:

- Existing scaffolding that previously mentioned pointers has been removed or will be migrated. The semantics and linter do not model raw pointers.
