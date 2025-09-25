# ami/stdlib/strings

Import: `ami/stdlib/strings`

Overview:
- Deterministic text primitives: Contains, HasPrefix/HasSuffix, Split/Join, Replace, Trim/TrimSpace, ToLower/ToUpper, Fields, Index/LastIndex, EqualFold.
- Pure functions (no I/O or globals); Unicode behavior is stable across platforms.

AMI Example (consistent with examples/correct/src/main.ami):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/strings >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline StringNormalize {
  Ingress(
    name=StringNormalize,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=string),
    worker=func(e Event<string>)(Event<string>, error){
      s := strings.TrimSpace(e.payload)
      s = strings.ToLower(s)
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
- EqualFold uses Unicode case folding; be aware of canonical equivalence rules. Example: EqualFold("Ångström","ångström") is true.
- Split on empty separator differs from splitting into runes; prefer explicit separators.
- TrimSpace removes all leading/trailing Unicode whitespace, not just ASCII.
