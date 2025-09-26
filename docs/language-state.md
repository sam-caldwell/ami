# Node-State Tables and Access (2.2.14, 2.3.5)

State is ambient and not passed as a parameter in worker functions:

- Preferred: `func f(ev Event<T>) (Event<U>, error)` and access state via `state.get/set/update/list`.
- Legacy `st State` parameter (third parameter) is deprecated but tolerated for now: `func f(ctx Context, ev Event<T>, st State) R`. Its use triggers `W_WORKER_STATE_PARAM_DEPRECATED`.
- The compiler may still surface `HasState` on worker references in pipeline IR for debugging during the transition.

Access Rules

- Writes must be explicitly marked: use `*stateField = value` to assign, or wrap state-mutating calls with `mutate(expr)`.
- AMI does not expose pointers; there is no address-of or dereference. Pointer `*State` parameters are rejected (`E_STATE_PARAM_POINTER`).
- State parameter is immutable (scaffold): reassignment is flagged (when detectable from tokens) as `E_STATE_PARAM_ASSIGN`.

Notes

- This is a token-level implementation suitable for early compiler scaffolding. A richer state-table model (with declaration of tables, key/value types, and controlled access APIs) can be layered on top of this by extending the AST and semantics.
