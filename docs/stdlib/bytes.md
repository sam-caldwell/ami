# ami/stdlib/bytes

Import: `ami/stdlib/bytes`

Overview:
- Deterministic byte-slice helpers: Contains, Compare, Index/LastIndex, Split/Join, Replace. Pure functions.

AMI Example (worker operates on base64 text converted to bytes):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/bytes >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline BytesContains {
  Ingress(
    name=BytesContains,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=string),
    worker=func(e Event<string>)(Event<bool>, error){
      b := []byte(e.payload)
      ok := bytes.Contains(b, []byte("needle"))
      return Event<bool>(ok)
    },
    minWorkers=1,maxWorkers=2,onError=ErrorPipeline,type=bool,
  ).Egress(
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=bool),
    worker=func(e Event<bool>){ stdio.Println(e.payload) },
    minWorkers=1,maxWorkers=1,onError=ErrorPipeline,capabilities=[],
  )
}
```

Edge Cases
- Compare returns sign of lexicographic comparison; use == 0 for equality.
- Replace with n < 0 replaces all occurrences.
