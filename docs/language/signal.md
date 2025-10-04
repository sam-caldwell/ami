# Language: signal (planned AMI example)

This document sketches how AMI source could use the stdlib `signal` package once language hooks are enabled. The host-backed Go implementation now resides under `src/ami/runtime/host/signal`; AMI sources live under `std/ami/stdlib/signal`.

AMI sample (illustrative only):

```
package app
import signal

func onInterrupt(){
  // cleanup and shutdown
}

func Main(){
  // Register handler for Ctrl+C / SIGINT
  signal.Register(SIGINT, onInterrupt)
  // ... continue work
}

pipeline P(){
  ingress;
  // Workers in this pipeline may rely on Main() having installed handlers
  egress;
}
```

Notes
- `signal.Register` takes a `SignalType` (e.g., `SIGINT`, `SIGTERM`) and a handler function.
- `signal.Enable`/`Disable` are optional, explicit OS hooks; `Register` enables OS delivery implicitly in the runtime.
- Advanced users can call `signal.Install` and `signal.Token` to preinstall handler thunks and exchange tokens deterministically.
