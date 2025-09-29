# Edge Specs (Stub)

This document describes compiler-generated edge specifications used by the AMI compiler. The current implementation
provides lightweight Go types to represent normalized edge configuration for analysis, IR emission, and future code
generation.

Kinds:

- FIFO: queue with optional bounds and backpressure. Backpressure in {block, dropOldest, dropNewest, shuntNewest,
shuntOldest}. Derived debug fields: bounded (Max>0), delivery (block=>atLeastOnce, else bestEffort).
- LIFO: stack with the same backpressure/bounds semantics as FIFO.
- Pipeline: reference to another pipeline by name and an optional textual payload type. Future phases will enforce
cross-pipeline type matching.
 - MultiPath: merge behavior configuration for Collect nodes. Supports simple k=v attributes and `merge.*(...)`
attribute calls. Deeper semantics (unknown attrs, conflicts, arg checks) are validated in the sem package. See
normalized merge scaffold in `contracts.v1`/`pipelines.v1` (`mergeNorm`) for Buffer/Stable/Sort.

These stubs are intentionally minimal to limit blast radius. They enable downstream phases to share a common structure
for edges without embedding parser/semantics details.

## edges.v1 (Debug Index)

When compiling with debug enabled, the compiler writes a per‑package edges index at:

`build/debug/asm/<pkg>/edges.json`

Schema:

- `schema`: fixed string `edges.v1`
- `package`: package name
- `edges`: array of edge entries with fields:
  - `unit`: compilation unit (basename without extension)
  - `from`, `to`: step names
  - `bounded`: bool derived from buffering attributes
  - `delivery`: `atLeastOnce` or `bestEffort` derived from backpressure policy
  - `type`: optional step type (from `type("T")` attribute)
  - `tinyBuffer`: bool hint when capacity is very small with drop policy
  - `collect` (optional): array of `edge.MultiPath` snapshots for `Collect` steps:
  - `unit`: unit name
  - `step`: step name where the multipath appears
  - `multipath`:
    - `args`: normalized argument list (e.g., input streams)
    - `merge`: list of merge attributes `{ name, args }`

Example:

```
{
  "schema": "edges.v1",
  "package": "app",
  "edges": [
    {"unit":"u","from":"ingress","to":"work","bounded":false,"delivery":"atLeastOnce"},
    {"unit":"u","from":"work","to":"egress","bounded":true,"delivery":"bestEffort","type":"X","tinyBuffer":true}
  ],
  "collect": [
    {
      "unit": "u",
      "step": "work",
      "multipath": {
        "args": ["inputA", "inputB"],
        "merge": [
          {"name":"merge.Sort", "args":["ts", "asc"]},
          {"name":"merge.Dedup", "args":["id"]}
        ]
      }
    }
  ]
}
```

## See Also
- `docs/backend/build_plan.md` — build plan may summarize debug artifact paths, including edges and IR indices.
- `docs/ir-indices.md` — IR indices emitted under `build/debug/ir/<pkg>/...` during verbose builds.

Example (Collect with multiple upstreams)

```
pipeline P(){
  A().Collect(merge.Buffer(4, dropOldest), merge.Stable(), merge.Sort(ts, asc)).B()
}
```

The debug snapshot includes raw `merge` attributes and a normalized `mergeNorm`:

```
"mergeNorm": {
  "buffer": {"capacity": 4, "policy": "dropOldest"},
  "stable": true,
  "sort": [{"field": "ts", "order": "asc"}]
}
```

## Concurrency Pragmas and pipelines.v1

Concurrency configuration is declared via pragmas and surfaced in `pipelines.v1` at the top level under `concurrency`.
Supported directives:

- `#pragma concurrency:workers N`: default worker pool size (N >= 1). Invalid values emit
`E_CONCURRENCY_WORKERS_INVALID`.
- `#pragma concurrency:schedule <policy>`: scheduling policy in `{fifo, lifo, fair, worksteal}`. Unknown values emit
`E_CONCURRENCY_SCHEDULE_INVALID`.
- `#pragma concurrency:limits ingress=N transform=N fanout=N collect=N mutable=N egress=N`: per-node-kind limits; each
value must be `>= 1`. Unknown keys emit `E_CONCURRENCY_LIMITS_KEY_UNKNOWN`; invalid values emit
`E_CONCURRENCY_LIMITS_INVALID`.

Example source:

```
#pragma concurrency:workers 4
#pragma concurrency:schedule fair
#pragma concurrency:limits ingress=2 transform=8 collect=4
pipeline P(){ ingress; work(); egress }
```

## Collect Instances and Edge IDs in pipelines.v1

When a pipeline contains multiple steps with the same name (e.g., multiple `Collect` steps), `pipelines.v1`
disambiguates instances and edges:

- Each step entry includes an `id` field (1‑based per step name in order of appearance).
- Each edge includes `fromId` and `toId` to reference specific step instances.
- For `Collect` steps, `multipath.inputs` lists upstream step names assigned to that specific instance.

Example excerpt:

```
{
  "pipelines": [
    {
      "name": "P",
      "steps": [
        {"name":"A","id":1},
        {"name":"Collect","id":1,"multipath":{"inputs":["A"]}},
        {"name":"B","id":1},
        {"name":"Collect","id":2,"multipath":{"inputs":["B"]}},
        {"name":"egress","id":1}
      ],
      "edges": [
        {"from":"A","fromId":1,"to":"Collect","toId":1},
        {"from":"B","fromId":1,"to":"Collect","toId":2}
      ]
    }
  ]
}
```

Connectivity metadata also includes occurrence‑level lists:

- `unreachableFromIngressIds`: array of `{name,id}` nodes not reachable from `ingress`.
- `cannotReachEgressIds`: array of `{name,id}` nodes that cannot reach `egress`.

pipelines.v1 header excerpt:

```
{
  "schema": "pipelines.v1",
  "package": "app",
  "unit": "u",
  "concurrency": {
    "workers": 4,
    "schedule": "fair",
    "limits": {"ingress": 2, "transform": 8, "collect": 4}
  },
  "pipelines": [
    {"name": "P", "steps": [ ... ]}
  ]
}
```

## Validation Examples

These small snippets illustrate common validation outcomes with their diagnostic codes.

- Undeclared node in an edge:
  - Source: `pipeline P(){ ingress; A; egress; A -> B; }`
  - Diag: `E_EDGE_UNDECLARED_TO` at `B`.

- Edge into `ingress` / out of `egress`:
  - Source: `pipeline P(){ ingress; A; egress; A -> ingress; egress -> A; }`
  - Diags: `E_EDGE_TO_INGRESS`, `E_EDGE_FROM_EGRESS`.

- Self edge and duplicate edge:
  - Source: `pipeline P(){ ingress; A; egress; A -> A; A -> egress; A -> egress; }`
  - Diags: `E_PIPELINE_SELF_EDGE` for `A -> A`, `W_PIPELINE_DUP_EDGE` for duplicate `A -> egress`.

- Disconnected node and no ingress→egress path:
  - Source: `pipeline P(){ ingress; A; B; egress; A -> B; }`
  - Diags: `E_PIPELINE_NODE_DISCONNECTED` (for `egress` if no incident edges), `E_PIPELINE_NO_PATH_INGRESS_EGRESS`.

- Unreachable from ingress / cannot reach egress:
  - Source: `pipeline P(){ ingress; A; B; egress; A -> egress; }`
  - Diags: `E_PIPELINE_UNREACHABLE_FROM_INGRESS` (for `B`), `E_PIPELINE_CANNOT_REACH_EGRESS` (when a node has no path to `egress`).

- Backpressure and capacity validation:
  - Source: `pipeline P(){ Collect edge.FIFO(min=10, max=5, backpressure=unknown); egress }`
  - Diags: `E_EDGE_CAPACITY_ORDER` (max < min), `E_EDGE_BACKPRESSURE`.
