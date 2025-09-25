# Memory Safety and “*” (AMI 2.3.2)

AMI does not expose raw pointers or process addresses. The `&` operator (address-of) is not part of the language, and unary `*` is not a dereference.

Usage of `*`:

- When `*` appears on the left-hand side of an assignment, it explicitly marks a mutating assignment in contexts where the compiler requires clarity. This does not imply pointer dereferencing.
- Outside this left-hand marking, `*` has no pointer semantics in AMI.

Compiler behavior:

- `&` anywhere in source is rejected (`E_PTR_UNSUPPORTED_SYNTAX`).
- Unary `*` in expression position is rejected (`E_STAR_MISUSED`).
- Types may include `*` for AMI-defined roles (e.g., `*State` in worker signatures). This does not introduce raw pointer operations.
- Mutability rules: AMI has no `mut { ... }` blocks. Mark assignments with `*` on the left-hand side (e.g., `*x = value`) or wrap side-effectful expressions in `mutate(expr)`.

Notes:

- Existing scaffolding that previously mentioned pointers has been removed or will be migrated. The semantics and linter do not model raw pointers.
