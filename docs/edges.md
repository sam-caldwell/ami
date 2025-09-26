# AMI Edge Specifications (`edge.*`)

Note: The authoritative source of truth is docs/Asynchronous Machine Interface.docx. If any guidance here conflicts with the .docx, the .docx prevails.

AMI programs declare edges inside node argument lists using compiler-generated constructs:

- `edge.FIFO(minCapacity, maxCapacity, backpressure, type)`
- `edge.LIFO(minCapacity, maxCapacity, backpressure, type)`
- `edge.Pipeline(name, minCapacity, maxCapacity, backpressure, type)`

These are not runtime calls. They are declarative specifications the compiler recognizes and lowers into concrete, high‑performance queue/linkage implementations during code generation.

Semantics (summary)
- Identity: `edge.Pipeline` bridges two pipelines by upstream ingress `name`.
- Capacity: `minCapacity`/`maxCapacity` bound internal buffers; `minCapacity >= 0` and `maxCapacity >= minCapacity`.
 - Backpressure: policies constrain enqueue behavior when buffers are full. Supported tokens:
  - `block`: producer blocks until capacity is available (maps to at‑least‑once delivery).
  - `dropOldest` and `dropNewest`: deterministically drop the oldest element (to admit the new one) or drop the incoming element, respectively (both map to best‑effort delivery).
  The bare token `drop` is deprecated/ambiguous; use `dropOldest` or `dropNewest` explicitly.
  The .docx is authoritative for behavior details.
- Type: specifies the payload type carried (`type=T`), enabling static analysis and specialization.
 - Bounded/unbounded: when `maxCapacity > 0` the edge is bounded; when `maxCapacity == 0` it is treated as unbounded.
 - Delivery guarantees (derived): `block` → `atLeastOnce`; `drop` → `bestEffort`.

Performance considerations
- Lock‑free or low‑contention rings: MPMC circular buffers sized to powers of two for cache‑friendly masking.
- Contiguous memory: pre‑allocated storage managed via RAII; AMI has no garbage collection. Capacity may grow within `[minCapacity,maxCapacity]` as configured.
- Zero‑copy handoff: pass event envelopes by reference; copy payloads only when required by isolation semantics.
- Cache locality: producer/consumer padding to prevent false sharing; per‑core shards if contention warrants.
- Backpressure fast paths: branch‑predictable states for `block` vs `drop`, minimal atomics on enqueue/dequeue.
- Pooling: bounded object pools for event envelopes and internal nodes to reduce allocations.

Compiler mapping
- Parser treats `edge.*(...)` as expressions in node argument lists.
- IR attaches an `edge.Spec` (see `src/ami/compiler/edge`) carrying the parameters.
- Codegen specializes queue implementations by payload type and selected policy.

Examples
- `Ingress(...).Transform( in=edge.FIFO(minCapacity=10, maxCapacity=20, backpressure=block, type=some.T), worker=workerFn, minWorkers=2, maxWorkers=8, type=some.U ).Egress(in=edge.FIFO(minCapacity=10, maxCapacity=20, backpressure=block, type=some.U))`
- `... Collect( in=edge.MultiPath(inputs=[ edge.FIFO(minCapacity=0, maxCapacity=8, backpressure=dropOldest, type=[]byte), edge.Pipeline(name=OtherPipe, minCapacity=64, maxCapacity=256, backpressure=block, type=[]byte) ], merge=Sort() ), minWorkers=1, maxWorkers=1, type=[]byte ).Egress(...)`

Expanded backpressure examples
- `Transform( in=edge.FIFO(minCapacity=64, maxCapacity=256, backpressure=dropOldest, type=bytes), worker=decode )`
- `Transform( in=edge.LIFO(minCapacity=0, maxCapacity=0, backpressure=dropNewest, type=Event), worker=dedupe )`
- `... Egress(in=edge.Pipeline(name=csvReaderPipeline, minCapacity=64, maxCapacity=256, backpressure=block, type=csv.Record), ...)`

Correctness notes (docx-aligned)
- Transform: declared with attributes including `in=edge.FIFO|edge.LIFO`, `worker=<func|factory>`, `minWorkers/maxWorkers`, and `type`. The worker is specified as `worker=MyFunc`, not as a positional argument.
- Collect: declared with `in=edge.MultiPath(...)` containing an `inputs=[ ... ]` list. The first entry is the default upstream edge (e.g., `edge.FIFO(...)`), followed by optional `edge.Pipeline(name=Upstream, ...)` entries. Merge behavior is expressed via `merge=Sort(...)` and related attributes.

Notes
- These specs live at compile time; they have no effect at runtime outside of the generated code. See the language docx §2.2.7/2.2.11 for background.
- AMI uses RAII for memory/resource management and does not have a garbage collector. Backpressure controls enqueue/dequeue behavior under capacity limits; it is unrelated to memory reclamation.

See also
- `docs/merge.md` for `Collect`-specific merge behavior configured via `edge.MultiPath(...)` and `merge.*(...)` attributes (sorting, stability, watermarking, buffering, and partitioning).

## MultiPath Scaffold Status

The current implementation supports `edge.MultiPath(...)` at a scaffold level:

- Parser accepts `in=edge.MultiPath(inputs=[...], merge=Name(args))` on `Collect` nodes.
- Semantics enforce basic rules (Collect-only, first input must be a default upstream edge, type compatibility across inputs, minimal merge-op name validation).
- IR/Schema: `pipelines.v1` includes MultiPath inputs and raw merge ops; `edges.v1` carries a debug snapshot. ASM listings emit `mp_*` pseudo-ops to aid testing and future integration.

For details and progress, see `SPECIFICATION.md` (sections 6.6 and 6.7) and `CONFLICTS.md` (edge.MultiPath section).

## Remaining Work (Merge Normalization)

Normalization and full validation of `merge.*` attributes are planned:

- Normalize merge configuration (key, sort {field, order, stable}, dedup, window, watermark {field, lateness}, timeout, buffer {capacity, backpressure}, partitionBy) into `pipelines.v1`.
- Validate per-attribute arity/types, detect conflicts, and enforce required fields.
- Map merge configuration to runtime orchestration with deterministic buffering and policy handling.
- Extend lints for merge smells (e.g., tiny buffers with drop policies, missing fields) and add golden tests for normalized IR.
