# Event and Error Typing & Contracts (1.7, 2.2)

This scaffold enforces core event/error typing rules and worker contracts:

Worker Contracts

- Canonical signature (ambient state): `func f(ev Event<T>) (Event<U>, error)`
  - State is ambient; use `state.get/set/update/list` rather than passing a `State` parameter.
  - Legacy 3-parameter form is not accepted for workers and is rejected with `E_WORKER_SIGNATURE`.
  - Invalid signatures emit `E_WORKER_SIGNATURE`.

Event Flow Contracts

- Across pipeline steps, upstream worker output payload `U` must match downstream worker input payload `T`:
  - If step i has workers producing `Event<U>`, and step i+1 has workers consuming `Event<T>`, then `U == T` is required.
  - Violations emit `E_EVENT_TYPE_FLOW`.
  - Error outputs (`Error<E>`) are ignored in normal-path flow checks.

Immutability

- Event parameter is immutable: assignments to the event parameter identifier are illegal (emits `E_EVENT_PARAM_ASSIGN`).
- Pointer/address safety (2.3.2): AMI does not expose raw pointers; `&` is not allowed (`E_PTR_UNSUPPORTED_SYNTAX`). Unary `*` is not a dereference and is only valid on the assignment LHS as a mutability marker.

Tests

- See `src/ami/compiler/sem/event_contracts_test.go` for event flow and parameter immutability.
- See `src/ami/compiler/sem/worker_test.go` for worker signature enforcement.
