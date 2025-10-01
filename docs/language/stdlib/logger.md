# Stdlib: logger (Pipelines and Sinks)

The `logger` stdlib package provides a buffered pipeline that batches writes to a sink with a configurable backpressure policy and optional JSON field redaction for `log.v1` records.

API (Go package `logger`)
- Backpressure policy: `type BackpressurePolicy` with values `Block`, `DropNewest`, `DropOldest`.
- `type Sink` interface: `Start() error`, `Write(line []byte) error`, `Close() error`.
- Built‑in sinks:
  - `NewFileSink(path string, perm os.FileMode) *FileSink`: append lines to a file, creating parents.
  - `NewStdoutSink(w io.Writer) *StdoutSink`, `NewStderrSink(w io.Writer) *StderrSink`.
- `type Config`:
  - `Capacity int`: channel capacity; `0` means synchronous handoff.
  - `BatchMax int`: flush when this many lines are buffered.
  - `FlushInterval time.Duration`: periodic flush; `0` disables time‑based flush.
  - `Policy BackpressurePolicy`: when the queue is full.
  - `Sink Sink`: destination.
  - Redaction (applies to `log.v1` JSON lines only):
    - `JSONAllowKeys []string`: optional allowlist (keep only these fields).
    - `JSONDenyKeys []string`: optional denylist (drop these fields).
    - `JSONRedactKeys []string`: redact exact keys to `"[REDACTED]"`.
    - `JSONRedactPrefixes []string`: redact any field with a matching key prefix.
- `NewPipeline(cfg Config) *Pipeline`: create an unstarted pipeline.
- `(*Pipeline).Start() error`: start background loop; initializes sink.
- `(*Pipeline).Enqueue(line []byte) error`: queue a line (policy‑dependent on full queue).
- `(*Pipeline).Close()`: stop loop, drain queue, flush remaining, and close sink.
- `type Stats { Enqueued, Written, Dropped, Batches, Flushes int64 }`; `(*Pipeline).Stats() Stats`.

Notes
- Time‑based flushing (`FlushInterval`) and size‑based flushing (`BatchMax`) can be combined.
- Redaction runs only when a line parses as a `log.v1` JSON record (see `schemas/log`); other lines bypass redaction.
- `DropNewest` and `DropOldest` increment the `Dropped` counter on policy‑based drops.

Examples
- File pipeline with batching and redaction
  ```go
  fs := logger.NewFileSink("build/debug/activity.log", 0)
  p := logger.NewPipeline(logger.Config{
      Capacity: 1024,
      BatchMax: 64,
      FlushInterval: 250 * time.Millisecond,
      Policy: logger.Block,
      Sink:   fs,
      JSONRedactKeys:     []string{"password"},
      JSONRedactPrefixes: []string{"secret.", "meta."},
  })
  _ = p.Start()
  _ = p.Enqueue([]byte("{\"schema\":\"log.v1\",\"timestamp\":\"...\",\"level\":\"info\",\"message\":\"ok\",\"fields\":{\"password\":\"x\",\"meta.token\":\"y\"}}\n"))
  p.Close()
  st := p.Stats()
  _ = st
  ```
- Stdout sink without batching
  ```go
  p := logger.NewPipeline(logger.Config{Capacity: 0, BatchMax: 1, Sink: logger.NewStdoutSink(nil)})
  _ = p.Start()
  _ = p.Enqueue([]byte("hello\n"))
  p.Close()
  ```

