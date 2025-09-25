# ami/stdlib/io

Import: `ami/stdlib/io`

Overview:
- Bounded copy primitives without implicit goroutines. API: CopyN(dst, src, n).

AMI Example (copy first N bytes of a stream):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/io >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline CopyFirstBytes {
  Ingress(
    name=CopyFirstBytes,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=ReaderWriterPair),
    worker=func(e Event<ReaderWriterPair>)(Event<uint64>, error){
      n, err := io.CopyN(e.payload.dst, e.payload.src, 1024)
      if err != nil { return nil, err }
      return Event<uint64>(n)
    },
    minWorkers=1,maxWorkers=2,onError=ErrorPipeline,type=uint64,
  ).Egress(
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=uint64),
    worker=func(e Event<uint64>){ stdio.Println(e.payload) },
    minWorkers=1,maxWorkers=1,onError=ErrorPipeline,capabilities=[],
  )
}
```

Note: `ReaderWriterPair` denotes a user-defined payload type carrying `src` and `dst` handles supplied by adjacent pipeline nodes.

Edge Cases
- If src has fewer than n bytes, CopyN returns the bytes copied and an error (short read).
