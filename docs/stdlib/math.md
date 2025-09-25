# ami/stdlib/math

Import: `ami/stdlib/math`

Overview:
- Deterministic constants (Pi, E, …) and core functions: Abs, Min/Max, Floor/Ceil/Trunc/Round, Pow, Sqrt. Pure functions.

AMI Example (numeric transform):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/math >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline MathOps {
  Ingress(
    name=MathOps,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=float64),
    worker=func(e Event<float64>)(Event<float64>, error){
      v := math.Round(e.payload * math.Pi)
      return Event<float64>(v)
    },
    minWorkers=1,maxWorkers=2,onError=ErrorPipeline,type=float64,
  ).Egress(
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=float64),
    worker=func(e Event<float64>){ stdio.Println(e.payload) },
    minWorkers=1,maxWorkers=1,onError=ErrorPipeline,capabilities=[],
  )
}
```

Edge Cases
- NaN propagates: comparisons with NaN yield NaN where applicable (e.g., Min/Max with NaN).
- Infinity: operations may return ±Inf; handle explicitly where required.
