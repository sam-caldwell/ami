# ami/stdlib/os

Import: `ami/stdlib/os`

Overview:
- Explicit filesystem I/O APIs with explicit permissions; no env/process mutation.

AMI Example (read then write transformed content):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/os >= v0.0.0
import ami/stdlib/strings >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline ReadWrite {
  Ingress(
    name=ReadWrite,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=string), // path string
    worker=func(e Event<string>)(Event<string>, error){
      data, err := os.ReadFile(e.payload)
      if err != nil { return nil, err }
      out := strings.ToUpper(string(data))
      err = os.WriteFile("/tmp/out.txt", []byte(out), 0o600)
      if err != nil { return nil, err }
      return Event<string>(out)
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
- Respect explicit perms (e.g., 0o600). Parent directories must exist for WriteFile unless created via Mkdir/MkdirAll.
- Stat errors on missing path; handle error paths deterministically.
