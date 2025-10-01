# Language: signal (planned AMI example)

This document sketches how AMI source could use the stdlib `signal` package once language hooks are enabled. The Go implementation (`src/ami/stdlib/signal`) is available today; this is forward‑looking AMI code.

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
- Handlers run sequentially when a matching signal arrives.
- See `docs/stdlib/signal.md` for the concrete Go API and behavior.
