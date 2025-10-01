# Language: time (planned AMI example)

This document sketches how AMI source could use the stdlib `time` package once language hooks are enabled. The Go implementation (`src/ami/stdlib/time`) is available today; this is forward‑looking AMI code.

AMI sample (illustrative only):

```
package app
import time

func Work(){
  // Sleep for 100 milliseconds
  time.Sleep(100ms)
  // Get current time and compute delta
  var t1 = time.Now()
  var t2 = time.Add(t1, 1s)
  var d = time.Delta(t1, t2) // 1s
  _ = d
}

pipeline P(){
  ingress;
  // Invoke Work() from workers as needed
  egress;
}
```

Notes
- `time.Sleep(d)` accepts a duration value (e.g., `100ms`, `1s`).
- `time.Now()` returns a `time.Time` value; `time.Add(t, d)` adds a duration; `time.Delta(t1, t2)` returns the difference.
- See `docs/stdlib/time.md` for the concrete Go API and behavior.
