# Stdlib: signal

The `signal` stdlib package provides basic signal handling.

API (Go package `amsignal`)
- `type SignalType`: enum-like type for common signals (`SIGINT`, `SIGTERM`, `SIGHUP`, `SIGQUIT`).
- `Register(sig SignalType, fn func())`: register a handler for a signal (multiple handlers allowed).
- `Reset()`: test helper to clear handlers and stop notifications.

Notes
- Handlers are invoked sequentially when a registered signal arrives.
- Platform differences:
  - On Windows, `SIGINT` maps to `os.Interrupt`. Other signals are best-effort.
  - On Unix-like systems, signals map to their syscall equivalents.
- Only catchable signals should be used with `Register`.

Examples
```go
amsignal.Register(amsignal.SIGINT, func(){
  // cleanup work
})
```

Tests
- See `src/ami/stdlib/signal/signal_test.go` for simple registration/dispatch checks.
