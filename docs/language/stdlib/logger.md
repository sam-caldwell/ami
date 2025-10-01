# Stdlib: logger (Pipelines and Sinks)

The `logger` module provides buffered pipelines that batch writes to a sink with a configurable backpressure policy and optional JSON field redaction for `log.v1` records.

API (AMI module `logger`)
- Backpressure policy constants: `logger.block`, `logger.dropNewest`, `logger.dropOldest`.
- Sinks:
  - `func logger.file(path string, perm int) Sink`
  - `func logger.stdout() Sink`, `func logger.stderr() Sink`
- `type PipelineConfig { capacity int, batchMax int, flushIntervalMs int, policy string, sink Sink, redact { allow slice<string>, deny slice<string>, keys slice<string>, prefixes slice<string> } }`
- `func logger.pipeline(cfg PipelineConfig) Pipeline`
- `method Pipeline.start() error`, `method Pipeline.enqueue(line bytes) error`, `method Pipeline.close()`
- `method Pipeline.stats() { enqueued int, written int, dropped int, batches int, flushes int }`

Notes
- Time‑based flushing (`FlushInterval`) and size‑based flushing (`BatchMax`) can be combined.
- Redaction runs only when a line parses as a `log.v1` JSON record (see `schemas/log`); other lines bypass redaction.
- `DropNewest` and `DropOldest` increment the `Dropped` counter on policy‑based drops.

Examples (AMI)
- File pipeline with batching and redaction
  ```
  import logger
  var p = logger.pipeline({
    capacity: 1024,
    batchMax: 64,
    flushIntervalMs: 250,
    policy: logger.block,
    sink: logger.file("build/debug/activity.log", 0o644),
    redact: { keys: ["password"], prefixes: ["secret.", "meta."] },
  })
  _ = p.start()
  _ = p.enqueue(stringToBytes("{\"schema\":\"log.v1\",\"timestamp\":\"...\",\"level\":\"info\",\"message\":\"ok\",\"fields\":{\"password\":\"x\",\"meta.token\":\"y\"}}\n"))
  p.close()
  var st = p.stats()
  ```
- Stdout sink without batching
  ```
  var p = logger.pipeline({capacity: 0, batchMax: 1, sink: logger.stdout()})
  _ = p.start()
  _ = p.enqueue(stringToBytes("hello\n"))
  p.close()
  ```
