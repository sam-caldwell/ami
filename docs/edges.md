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

Concurrency configuration is declared via pragmas and surfaced in `pipelines.v1` at the top level under
`concurrency`. Supported directives:

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
