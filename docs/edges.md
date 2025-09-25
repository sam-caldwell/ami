# AMI Edge Specifications (`edge.*`)

AMI programs declare edges inside node argument lists using compiler-generated constructs:

- `edge.FIFO(minCapacity, maxCapacity, backpressure, type)`
- `edge.LIFO(minCapacity, maxCapacity, backpressure, type)`
- `edge.Pipeline(name, minCapacity, maxCapacity, backpressure, type)`

These are not runtime calls. They are declarative specifications the compiler recognizes and lowers into concrete, high‑performance queue/linkage implementations during code generation.

Semantics (summary)
- Identity: `edge.Pipeline` bridges two pipelines by upstream ingress `name`.
- Capacity: `minCapacity`/`maxCapacity` bound internal buffers; `minCapacity >= 0` and `maxCapacity >= minCapacity`.
- Backpressure: `block` or `drop` (initial set). Additional policies may be added in later revisions.
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
- `Ingress(...).Transform(...).Egress(in=edge.FIFO(minCapacity=10, maxCapacity=20, backpressure=block, type=some.T), ...)`
- `... Collect().Egress(in=edge.LIFO(minCapacity=0, maxCapacity=8, backpressure=drop, type=[]byte), ...)`
- `... Egress(in=edge.Pipeline(name=csvReaderPipeline, minCapacity=64, maxCapacity=256, backpressure=block, type=csv.Record), ...)`

Notes
- These specs live at compile time; they have no effect at runtime outside of the generated code. See the language docx §2.2.7/2.2.11 for background.
- AMI uses RAII for memory/resource management and does not have a garbage collector. Backpressure controls enqueue/dequeue behavior under capacity limits; it is unrelated to memory reclamation.
