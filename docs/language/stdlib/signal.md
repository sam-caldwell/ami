# Stdlib: signal

The `signal` module provides basic process signal handling for AMI programs.

API (AMI module `signal`)
- `type SignalType` — enum‑like type for common signals (`SIGINT`, `SIGTERM`, `SIGHUP`, `SIGQUIT`).
- `func signal.register(sig SignalType, fn func())` — register a handler (multiple handlers allowed).
- `func signal.reset()` — test helper to clear handlers and stop notifications.

Notes
- Handlers are invoked sequentially when a registered signal arrives.
- Platform differences:
  - On Windows, `SIGINT` maps to `os.Interrupt`. Other signals are best-effort.
  - On Unix-like systems, signals map to their syscall equivalents.
- Only catchable signals should be used with `Register`.

Examples (AMI)
```
import signal

signal.register(SIGINT, func(){
  // cleanup work
})
```

Tests
- See `src/ami/stdlib/signal/signal_test.go` for simple registration/dispatch checks.
