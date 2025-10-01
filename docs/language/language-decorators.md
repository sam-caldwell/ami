# AMI Function Decorators

Overview

- Decorators annotate functions with metadata: `@name` or `@name(arg1, arg2, ...)`.
- They apply to functions only; not allowed on pipelines/structs/enums.
- Order is bottom-to-top (Python semantics) and preserved in AST and IR debug.

Built-ins

- `@deprecated("message")`: emits a `W_DEPRECATED` diagnostic with the provided message.
- `@metrics`: reserved for observability; no-op in current scaffold.

Resolution

- Name resolves to a known built-in or a top-level function in the same file.
- Unknown/undefined emits `E_DECORATOR_UNDEFINED`.
- Conflicting duplicate decorator arguments on the same function emit `E_DECORATOR_CONFLICT`.
- Decorators can be disabled via configuration; disabled names emit `E_DECORATOR_DISABLED`.

Workers

- Workers must not be decorated; using a decorated function as a worker emits `E_DECORATOR_ON_WORKER`.
- Decorators must not change a worker’s externally visible signature. If a decorated function is referenced as a worker and the signature is not `func(Event<T>) (Event<U>, error)`, emit `E_DECORATOR_SIGNATURE`.

Configuration

- Tooling may disable specific decorators via workspace configuration (scaffold): the analyzer exposes `SetDisabledDecorators(...)` for tests and integration wiring.

Determinism

- Decorator lists are preserved in source order in AST and IR debug outputs; diagnostics include ISO-8601 UTC timestamps.

Examples

```
@deprecated("use G2 instead")
func G() {}

@metrics
func H(ev Event<T>) (Event<U>, error) { return ev, nil } // E_DECORATOR_ON_WORKER
```

