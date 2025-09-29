# Runtime KV Store Overview

The runtime key/value store under `src/ami/runtime/kvstore` provides a simple, concurrency-safe, in-memory state
mechanism used by the runtime harness and tests. It supports namespacing, TTL, delete-on-read, and optional LRU
eviction, making it useful for attaching ephemeral state to pipelines or nodes during execution and testing.

## Design

- Default store: process-global instance with convenience API (`kvstore.Put/Get/Del/Has/Keys/Stats/SetCapacity`).
- Namespaces: `kvstore.Namespace("<ns>")` returns a per-namespace `*Store` (e.g., `pipeline/node`).
- Concurrency: all operations are safe across goroutines.
- TTL: entries can expire after a duration; sliding TTL refreshes on successful `Get`.
- Delete-on-read: `WithMaxReads(n)` decrements on `Get` and deletes when reads reach zero.
- Capacity/LRU: optional cap evicts least-recently used entries (front eviction) when exceeded.
- Metrics: `Hits`, `Misses`, `Expirations`, `Evictions`, `CurrentSize` (via `Store.Metrics()` or `kvstore.Stats()`).

## API Highlights

- `kvstore.Put(key, val, opts...)`: store value in default store.
  - Options: `WithTTL(d)`, `WithSlidingTTL()`, `WithMaxReads(n)`.
- `kvstore.Get(key) (any, bool)`: fetch value if present and not expired; applies sliding TTL and max-reads.
- `kvstore.Del(key) bool`, `kvstore.Has(key) bool`, `kvstore.Keys() []string`.
- `kvstore.Default() *Store`, `kvstore.Namespace(ns) *Store` for explicit store usage.
- `(*Store).SetCapacity(n int)`: enable LRU eviction for that store; `kvstore.SetCapacity(n)` for default.

## Integration Points

- Build (verbose): `ami build --verbose` writes process-level KV artifacts under `build/debug/kv/`:
  - `metrics.json` (`kv.metrics.v1`)
  - `dump.json` (`kv.dump.v1`, keys only)
  Source: `src/cmd/ami/build_run.go`.

- Test harness: `ami test` supports KV controls and artifacts (see `docs/runtime-tests.md`).
  - Flags: `--kv-metrics`, `--kv-dump`, `--kv-events` (JSON diag stream), and per-case `emit=true`.
  - Per-case artifacts (when `emit=true` or `--verbose`): `build/test/kv/<file>_<case>.(metrics|dump).json`.
  - Namespaces: per-case `#pragma test:kv ns="..."` selects a namespaced store.
  Source: `src/cmd/ami/runtime_exec.go` and `src/ami/runtime/tester`.

## Usage Notes

- Treat the store as ephemeral: it is meant for transient coordination, not durable state.
- Prefer namespacing (`pipeline/node`) to avoid key collisions across tests or concurrent runs.
- Sliding TTL is useful for session-like entries; delete-on-read is useful for one-shot tokens.
- Capacity defaults to unbounded; set caps in long-running scenarios to bound memory and observe `Evictions`.

