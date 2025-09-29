# AMI Events: Lifecycle and Immutability

Overview

- Events are immutable payload carriers flowing through pipelines.
- Every event includes metadata fields: `id` (stable UUID), `timestamp` (ISO‑8601 UTC), `attempt` (int), and `trace`
  context (`traceparent`, `tracestate`).
- Payload immutability is enforced by the language: workers receive `Event<T>` and must return a new `Event<U>`; the
  input payload is not mutated in place.

Constraints

- No raw pointers in event payload types: event type parameters (`T`, `U`) must be pointer‑free.
- Address‑of (`&`) is prohibited in AMI. Unary `*` is reserved for the mutating assignment marker on the left‑hand
  side
  (`*name = expr`).
- Workers follow the canonical signature: `func(ev Event<T>) (Event<U>, error)`. Factories `New*` are exempt.

Determinism

- All event timestamps and diagnostic records use ISO‑8601 UTC string format with millisecond precision.
- JSON output mode disables ANSI colors. Human vs JSON renders remain consistent in content.

Notes

- Compiler/IR emit `eventmeta.v1` in debug builds with `{id,timestamp,attempt,trace,immutablePayload:true}` for each
  unit, capturing the contract.
- Additional body‑level immutability checks will arrive with the Imperative subset landing (mutability markers, RAII,
  and ownership flow).

