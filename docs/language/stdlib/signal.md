# Stdlib: signal

The `signal` module exposes deterministic, language‑level signal registration and OS hooks.

API (AMI module `signal`)
- `type SignalType` — enum for common signals: `SIGINT`, `SIGTERM`, `SIGHUP`, `SIGQUIT`.
- `func Register(sig SignalType, fn any)` — register a handler for `sig`.
- `func Install(fn any)` — install a handler thunk for `fn` (advanced; optional).
- `func Token(fn any) int64` — deterministic token for `fn` (advanced; optional).
- `func Enable(sig SignalType)` — enable OS delivery for `sig` (optional; `Register` enables it implicitly at runtime).
- `func Disable(sig SignalType)` — disable OS delivery for `sig` (optional).

Behavior
- Handlers are invoked when a matching signal arrives; order is not guaranteed.
- Handler identity is represented by a deterministic token (FNV‑1a of the handler name or a stable fallback).
- `Register` persists the handler token and enables OS delivery for the signal.
- `Install/Token` let advanced users preinstall or exchange handler thunks via tokens; not required for typical cases.
- On POSIX builds (`-tags runtime_posix`), OS enablement wires a process trampoline with `signal(2)` to route to the registered handler.

Signal Mapping
- `SIGINT=2`, `SIGTERM=15`, `SIGHUP=1`, `SIGQUIT=3` (POSIX mapping).
- Only catchable signals should be used.

Examples
- Register a handler
```
import signal

func onInterrupt(){
  // cleanup and shutdown
}

func main(){
  signal.Register(SIGINT, onInterrupt)
}
```

- Explicitly enable/disable OS delivery
```
import signal

func main(){
  signal.Enable(SIGTERM)   // optional; Register calls Enable internally at runtime
  // ...
  signal.Disable(SIGTERM)  // optional: revert to default behavior
}
```

- Advanced: preinstall a thunk and share a token
```
import signal

func H(){}

func setup(){
  // Install H's thunk in the local module (token is deterministic)
  signal.Install(H)
  // Persist or send the token if another module needs to look it up
  let tok = signal.Token(H)
  // ... use tok in cross‑module coordination
}
```

Notes
- Cross‑module function pointers are not taken implicitly; `Install` for a selector like `m.H` installs only the token.
- On POSIX, the runtime installs a single trampoline per signal; the handler thunk dispatch uses the stored token.
