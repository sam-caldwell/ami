# Node-State Tables and Access (2.2.14, 2.3.5)

This scaffold models per-node state as an explicit capability argument in worker functions (no pointer syntax):

- Worker signature includes `st State` (third parameter): `func f(ctx Context, ev Event<T>, st State) R`.
- The compiler captures `HasState` on worker references in pipeline IR for debugging/analysis.

Access Rules

- Writes must occur inside `mut { ... }` blocks (enforced by existing mutability rule `E_MUT_ASSIGN_OUTSIDE`).
- AMI does not expose pointers; there is no address-of or dereference for `st`.
- State parameter is immutable (scaffold): reassignment is flagged (when detectable from tokens) as `E_STATE_PARAM_ASSIGN`.

Notes

- This is a token-level implementation suitable for early compiler scaffolding. A richer state-table model (with declaration of tables, key/value types, and controlled access APIs) can be layered on top of this by extending the AST and semantics.
