# Node-State Tables and Access (2.2.14, 2.3.5)

This scaffold models per-node state via an explicit `*State` parameter on worker functions:

- Worker signature includes `st *State` (third parameter): `func f(ctx Context, ev Event<T>, st *State) R`.
- The compiler captures `HasState` on worker references in pipeline IR for debugging/analysis.

Access Rules

- Writes must occur inside `mut { ... }` blocks (enforced by existing mutability rule `E_MUT_ASSIGN_OUTSIDE`).
- Pointer safety applies to `st` like any other pointer: unsafe dereference outside a `st != nil` guard emits `E_DEREF_UNSAFE`.
- State parameter is immutable (scaffold): reassignment and address-of of the state parameter are flagged (when detectable from tokens) as `E_STATE_PARAM_ASSIGN` and `E_STATE_ADDR_PARAM`.

Notes

- This is a token-level implementation suitable for early compiler scaffolding. A richer state-table model (with declaration of tables, key/value types, and controlled access APIs) can be layered on top of this by extending the AST and semantics.

