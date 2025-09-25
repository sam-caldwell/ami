# Memory Safety and “*” (AMI 2.3.2)

AMI does not expose raw pointers or process addresses. The `&` operator (address-of) is not part of the language, and unary `*` is not a dereference.

Usage of `*`:

- When `*` appears on the left-hand side of an assignment, it explicitly marks a mutating assignment in contexts where the compiler requires clarity. This does not imply pointer dereferencing.
- Outside this left-hand marking, `*` has no pointer semantics in AMI.

Compiler behavior:

- Any attempt to use pointer syntax (e.g., `&x` or `*x` as a dereference, or types like `*T`) is rejected during parsing with `E_PTR_UNSUPPORTED_SYNTAX`.
- Mutability rules still apply: writes must appear in `mut { ... }` blocks per the language rules.

Notes:

- Existing scaffolding that previously mentioned pointers has been removed or will be migrated. The semantics and linter do not model raw pointers.

