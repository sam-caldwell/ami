# Data Mutability (AMI 2.3.x)

AMI uses explicit, local mutability with no pointer semantics:

- Mutate marker: use `*` on the left-hand side to mark a mutating assignment, e.g., `*x = value`.
- Atomic mutate: wrap expressions in `mutate(<expression>)` to request an atomic, VM-managed mutation consistent with mutability rules.
- No pointers: unary `*` is not a dereference, and `&` is not allowed (see Memory Safety, AMI 2.3.2).

Rules

- Marking required: unmarked assignment (`x = y`) emits `E_MUT_ASSIGN_UNMARKED`.
- LHS-only `*`: using unary `*` outside an assignment LHS is invalid.
- Size-match rule: mutations must preserve container size/shape where required by the VM. Examples (from the spec):
  - `*b = a    // allowed` when sizes match
  - `*c = a    // fails` when sizes do not match
  - `*a = []byte{05,06,07,08} // allowed` when sizes match
- Function boundaries: unless otherwise specified by the spec, mutability does not traverse function boundaries; `mutate(…)` and `*` follow normal argument/result isolation rules.

Notes

- There are no Rust-like `mut { ... }` blocks in AMI.
- The semantic analyzer enforces the mutate marker and size‑match rules and rejects any legacy pointer-style usage.

Examples

```
*x = y                      // marked assignment
result = mutate(f(a)+g(b))  // atomic mutate of composed expr
```
