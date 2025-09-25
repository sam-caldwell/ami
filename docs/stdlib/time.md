# ami/stdlib/time

Import: `ami/stdlib/time`

Overview:
- Deterministic time: ISO‑8601 UTC with milliseconds; Duration parse/format; injected Clock for Now.

AMI Example (format timestamps to RFC‑3339 UTC millis):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/time >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline TimeFormat {
  Ingress(
    name=TimeFormat,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=time.Time),
    worker=func(e Event<time.Time>)(Event<string>, error){
      s := time.FormatRFC3339Millis(e.payload)
      return Event<string>(s)
    },
    minWorkers=1,maxWorkers=2,onError=ErrorPipeline,type=string,
  ).Egress(
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=string),
    worker=func(e Event<string>){ stdio.Println(e.payload) },
    minWorkers=1,maxWorkers=1,onError=ErrorPipeline,capabilities=[],
  )
}
```

Edge Cases
- Parsing requires millisecond precision; strings must match the exact layout: `YYYY-MM-DDThh:mm:ss.mmmZ`.
- Durations follow Go’s `time.ParseDuration` canonical form; format emits canonical strings, e.g. `1h2m3.004s`.
- No ambient time reads; use injected clocks for Now in runtimes.
