# ami/stdlib/encoding/json

Import: `ami/stdlib/encoding/json`

Overview:
- Deterministic JSON: Map/object keys are sorted lexicographically. APIs: Marshal, MarshalIndent, Unmarshal, UnmarshalStrict, Canonicalize, Compact, Valid, EqualCanonical.

AMI Example (canonicalize, then strict decode):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/encoding/json >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline CanonicalJSON {
  Ingress(
    name=CanonicalJSON,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=string), // raw JSON
    worker=func(e Event<string>)(Event<string>, error){
      // canonicalize and re-emit
      b, err := json.Canonicalize([]byte(e.payload))
      if err != nil { return nil, err }
      return Event<string>(string(b))
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
- Strict decode rejects unknown fields; prefer UnmarshalStrict for schema enforcement.
- Numeric values use UseNumber by default in Unmarshal helpers to avoid float rounding surprises.
