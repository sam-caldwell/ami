# Node-State Table (Ephemeral Key Store)

This runtime package provides an in-memory, per-node, per-pipeline key/value store with TTLs, delete-on-read, LRU eviction, and metrics.

Key properties

- Ephemeral: process-local; cleared on restart.
- Namespaced: a `Registry` yields isolated stores per `(pipeline,node)`.
- Concurrency-safe: all operations are atomic at the key level.
- TTLs: absolute or sliding deadlines; optional background sweeper.
- Delete-on-read: remove after N reads (one-time reads supported).
- Capacity limits: approximate memory cap with LRU eviction when exceeded.
- Metrics: hits, misses, expirations, evictions, current entries and bytes.
- Observability: `DebugDump()` returns a JSON summary suitable for `--verbose` logging.

API overview

- `store := kvstore.New(options)` create a store.
- `put(key, val[, TTL(d), SlidingTTL(d), MaxReads(n)])`
- `get(key) -> (val, ok)`
- `del(key) -> bool`
- `has(key) -> bool`
- `keys() -> []string`
- `metrics() -> Stats`
- `DebugDump() -> string` (emit under `--verbose` if desired)

Namespacing and wiring

- Use `kvstore.NewRegistry(opts)` to obtain a `Registry` for your runtime instance.
- Call `reg.Get(pipeline, node)` to get an isolated store for that node.
- A process-wide default registry is available via `kvstore.Default()` for scaffolding.

Guarantees and limitations

- Memory cap is approximate; sizes are estimated via JSON encoding or length for strings/bytes.
- TTL enforcement occurs on access and via an optional background sweep. With `SweepInterval<=0`, expiry happens lazily when the key is next accessed.
- Sliding TTL refresh occurs on successful `Get()`.
- Delete-on-read deletes after the Nth successful `Get()` and still returns the value for that final read.
- Values are stored as opaque interfaces; callers are responsible for type assertions on `Get()`.

