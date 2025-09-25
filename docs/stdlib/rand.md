# ami/stdlib/rand

Import: `ami/stdlib/rand`

Overview:
- Deterministic PRNG with explicit seed: New(seed). Methods: Intn, Uint64, Read, Perm. No globals.

AMI Example (seeded random ints):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/rand >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline DeterministicRand {
  Ingress(
    name=DeterministicRand,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=int),
    worker=func(e Event<int>)(Event<int>, error){
      r := rand.New(42)
      return Event<int>( r.Intn(100) )
    },
    minWorkers=1,maxWorkers=2,onError=ErrorPipeline,type=int,
  ).Egress(
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=int),
    worker=func(e Event<int>){ stdio.Println(e.payload) },
    minWorkers=1,maxWorkers=1,onError=ErrorPipeline,capabilities=[],
  )
}
```

Edge Cases
- Intn panics if n <= 0; validate inputs upstream.
- Read fills the buffer fully or returns an error.
