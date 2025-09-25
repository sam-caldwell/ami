# `defer`

Status: AMI language feature. Semantics match Go’s `defer` from the user’s perspective.

- Syntax: `defer <call-expr>` inside function bodies.
- Argument evaluation: arguments to the deferred call are evaluated immediately at the point of the `defer` statement.
- Execution order: deferred calls execute in last‑in, first‑out (LIFO) order when the surrounding function returns.
- Return/abnormal exit: deferred calls run on normal return and during unwinding due to faults/panics; multiple defers still run in LIFO order.
- Interaction with results: deferred functions execute after the function’s return values are set; they may observe or modify named result variables before return is finalized (same as Go).
- Memory safety: `defer` does not introduce pointer semantics; it composes with `mutate(...)` and the LHS `*` mutation marker rules.

Examples:

```
func f(ctx Context, ev Event<string>, st State) Event<string> {
    res := openResource()
    defer res.Close()         // release at end of function
    // ... res is still usable here ...
    return ev
}

func g(ctx Context, ev Event<string>, st State) Event<string> {
    h := acquireHandle()
    defer release(h)          // function-style release at exit
    return ev
}
```
