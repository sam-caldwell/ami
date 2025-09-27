# Edge Specs (Stub)

This document describes compiler-generated edge specifications used by the AMI compiler. The current implementation provides lightweight Go types to represent normalized edge configuration for analysis, IR emission, and future code generation.

Kinds:

- FIFO: queue with optional bounds and backpressure. Backpressure in {block, dropOldest, dropNewest, shuntNewest, shuntOldest}. Derived debug fields: bounded (Max>0), delivery (block=>atLeastOnce, else bestEffort).
- LIFO: stack with the same backpressure/bounds semantics as FIFO.
- Pipeline: reference to another pipeline by name and an optional textual payload type. Future phases will enforce cross-pipeline type matching.
- MultiPath: merge behavior configuration for Collect nodes. Supports simple k=v attributes and `merge.*(...)` attribute calls. Deeper semantics (unknown attrs, conflicts, arg checks) are validated in the sem package.

These stubs are intentionally minimal to limit blast radius. They enable downstream phases to share a common structure for edges without embedding parser/semantics details.

