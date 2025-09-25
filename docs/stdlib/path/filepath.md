# ami/stdlib/path/filepath

Import: `ami/stdlib/path/filepath`

Overview:
- Deterministic, pure string path operations: Clean, Join, Base/Dir, Ext. No symlink evaluation.
- All separators normalized to '/'.

AMI Example (normalize paths):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/path/filepath >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline NormalizePath {
  Ingress(
    name=NormalizePath,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=string),
    worker=func(e Event<string>)(Event<string>, error){
      p := filepath.Clean(e.payload)
      return Event<string>(p)
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
- `..` and `.` handled lexically; no filesystem access.
- Windows-style `\` is normalized to `/` prior to processing for stable results.
