# Merge Attributes (edge.MultiPath on Collect)

This document describes the `merge.*` attributes that can be specified within
`edge.MultiPath(...)` on `Collect` nodes.

Syntax examples:

- edge.MultiPath(inputs=[ ... ], merge=Sort("ts","asc"))
- edge.MultiPath(inputs=[ ... ], merge=Sort("ts","desc"), merge=Stable(), merge=Buffer(10,dropOldest))

Normalized configuration captured in debug schemas:

- pipelines.v1: `inEdge.multiPath.mergeConfig`
- edges.v1: `multiPath.mergeConfig`

Attributes (tolerant parser):

- merge.Sort(field[, order]): order in {asc, desc}; default asc
- merge.Stable(): request stable sorting for tiebreaks
- merge.Key(field): key selector for other ops
- merge.Dedup([field]): deduplicate by optional field (defaults to key)
- merge.Window(size): in-flight merge window size (>0)
- merge.Watermark(field, lateness): lateness captured as a tolerant string
- merge.Timeout(ms): timeout in milliseconds (int)
- merge.Buffer(capacity[, backpressure]): backpressure in {block, dropOldest, dropNewest}
- merge.PartitionBy(field): partition streams prior to merging

Diagnostics and hints:

- E_MERGE_ATTR_UNKNOWN, E_MERGE_ATTR_ARGS, E_MERGE_ATTR_REQUIRED, E_MERGE_ATTR_CONFLICT (reserved)
- W_MERGE_SORT_NO_FIELD, W_MERGE_TINY_BUFFER, W_MERGE_WATERMARK_MISSING_FIELD, W_MERGE_WINDOW_ZERO_OR_NEGATIVE

Notes:

- Attributes are order-insensitive; last-write-wins for duplicates.
- Normalization is scaffolded; runtime operator planning is deferred.

