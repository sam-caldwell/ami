# ami/stdlib/regexp

Import: `ami/stdlib/regexp`

Overview:
- Deterministic subset of RE2 regex. APIs: Compile/MustCompile, MatchString. No global caches.

AMI Example (validate numeric input):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/regexp >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline OnlyDigits {
  Ingress(
    name=OnlyDigits,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=string),
    worker=func(e Event<string>)(Event<bool>, error){
      r := regexp.MustCompile("^[0-9]+$")
      return Event<bool>( r.MatchString(e.payload) )
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
- Invalid or potentially catastrophic patterns may be rejected at Compile time.
- Anchors ^ and $ operate in multi-line mode only if explicitly designed; prefer explicit patterns.
