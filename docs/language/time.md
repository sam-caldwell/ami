# Language: time (planned AMI example)

This document shows how AMI source uses the stdlib `time` package. The runtime implementation is opaque to AMI users; the API here is the stable AMI surface.

AMI sample (illustrative only):

```
package app
import time

func Work(){
  // Sleep for 100 milliseconds
  time.sleep(100ms)
  // Get current time and compute delta
  let t1 = time.now()
  let t2 = time.add(t1, 1s)
  let d = time.delta(t1, t2) // 1s
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
- See `docs/language/stdlib/time.md` for the concrete Go API and behavior.
