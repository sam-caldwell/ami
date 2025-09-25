# ami/stdlib/bufio

Import: `ami/stdlib/bufio`

Overview:
- Buffered I/O with explicit buffer size; Flush semantics are explicit and tested. API: NewReaderSize, NewWriterSize, Read, Write, Flush.

AMI Example (buffered write with explicit flush):

```ami
// file: main.ami
package main:0.0.1

import ami/stdlib/bufio >= v0.0.0
import github.com/asymmetric-effort/ami/stdio >= v0.0.0

pipeline BufferedWrite {
  Ingress(
    name=BufferedWrite,
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=WriterPair),
    worker=func(e Event<WriterPair>)(Event<bool>, error){
      w, err := bufio.NewWriterSize(e.payload.w, 4096)
      if err != nil { return nil, err }
      _, _ = w.Write([]byte("hello"))
      if err := w.Flush(); err != nil { return nil, err }
      return Event<bool>(true)
    },
    minWorkers=1,maxWorkers=2,onError=ErrorPipeline,type=bool,
  ).Egress(
    in=edge.FIFO(minCapacity=1,maxCapacity=8,backpressure=block,type=bool),
    worker=func(e Event<bool>){ stdio.Println(e.payload) },
    minWorkers=1,maxWorkers=1,onError=ErrorPipeline,capabilities=[],
  )
}
```

Note: `WriterPair` denotes a user-defined payload type carrying an underlying `w` handle provided by an adjacent pipeline node.

Edge Cases
- Size must be > 0 or construction fails.
- Buffer writes may not reach underlying writer until buffer is full or Flush is called.

Writer Flush Semantics
- Write buffering: `Write(p)` copies into the internal buffer up to its capacity.
- Partial flush on full buffer: If a `Write` fills the buffer and additional input remains, it performs an immediate flush to the underlying writer before continuing. This makes partial output visible right away when input exceeds the buffer size.
- Final flush: Remaining buffered data is written to the underlying writer when `Flush()` is called (or by a subsequent `Write` that fills and triggers another partial flush).
- Errors: Any underlying flush error is returned; `Write` reports bytes written up to the error.

Example
```
// buffer size = 3, write "hello"
// After Write: underlying writer sees "hel" (partial flush), remainder "lo" buffered
// After Flush: underlying writer sees "hello"
```

Reader Semantics (brief)
- Read buffering: `Read(p)` serves from an internal buffer first; if empty, it refills up to the configured capacity from the underlying reader.
- Return values: may return fewer than `len(p)` bytes even when not at EOF; subsequent calls continue from the internal buffer.
- EOF: `Read` returns `io.EOF` only after the internal buffer is drained and the underlying reader reports end of input.
- Errors: underlying read errors are propagated. If bytes were read before an error, those bytes are returned with the error on a subsequent call once buffered data is drained.

Example
```
// buffer size = 2, source = "abcdef"
// First Read(4): may return "abcd" (4 bytes) assembled from two internal refills
// Next Read(4): returns "ef" then io.EOF on subsequent call
```
