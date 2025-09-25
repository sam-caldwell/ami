# Memory Model: Ownership & RAII (2.4)

This scaffold introduces a minimal ownership/RAII discipline for AMI consistent with the spec’s RAII model.

About `Owned<T>` (scaffold marker)

- `Owned<T>` is an internal analysis marker used by lints/IR in this repository; it is not a source-level AMI type.
- The linter treats parameters modeled as `Owned<T>` as requiring an explicit handoff or release before function end.

Rules

- Release or transfer: Each `Owned<T>` parameter must be either released or transferred before the function ends.
  - Release: calling `release(x)`, `drop(x)`, `free(x)`, or `dispose(x)`; or method-like `x.Close()`, `x.Release()`, `x.Free()`, `x.Dispose()`.
  - Transfer: passing `x` to a function whose corresponding parameter type is `Owned<…>`.
- Double release: multiple releases/transfers for the same variable emit `E_RAII_DOUBLE_RELEASE`.
- Use after release: any subsequent use of the variable emits `E_RAII_USE_AFTER_RELEASE`.

Scope & Limitations

- Token-based: The checker scans function body tokens; it does not model local variable declarations or return values. It only tracks function parameter ownership.
- Results and returns: Returning ownership is not modeled; prefer release/transfer in this scaffold.

Tests

- See `src/ami/compiler/sem/raii_semantics_test.go` for happy/sad tests.
