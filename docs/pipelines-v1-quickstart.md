# pipelines.v1 Quickstart

This guide explains how to read the `pipelines.v1` debug artifact emitted during compilation. It summarizes a package’s
pipelines, steps, edges, and concurrency metadata to make validation and visualization easier.

## Where to find it

- Written when debug/verbose artifacts are enabled, under `build/debug/asm/<pkg>/pipelines.json` (see compiler wiring).

## Top-level header

- `schema`: fixed string `pipelines.v1`.
- `package`: package name.
- `unit`: compilation unit (basename without extension) to disambiguate multiple files per package.
- `concurrency`:
  - `workers`: default worker pool size (`>=1`).
  - `schedule`: `fifo|lifo|fair|worksteal`.
  - `limits`: object with per-node-kind caps: `ingress|transform|fanout|collect|mutable|egress`.

Example header:

```
{
  "schema": "pipelines.v1",
  "package": "app",
  "unit": "u",
  "concurrency": {"workers": 4, "schedule": "fair", "limits": {"ingress": 2, "transform": 8}}
}
```

## Pipelines array

Each entry describes one pipeline:

- `name`: pipeline name.
- `steps`: ordered list of step entries with occurrence `id`:
  - `name`: step name (`ingress`, `egress`, `Collect`, etc.).
  - `id`: 1-based per name, in order of appearance (disambiguates repeated names).
  - `type` (optional): textual type from `type("...")` attribute.
  - `multipath` (Collect only): normalized merge configuration; includes `inputs` listing upstream step names for this instance.
- `edges`: list of connections with instance-aware references:
  - `from`, `to`: step names.
  - `fromId`, `toId`: specific occurrence IDs when multiple instances exist.

Minimal example:

```
{
  "pipelines": [
    {
      "name": "P",
      "steps": [
        {"name":"ingress","id":1},
        {"name":"A","id":1},
        {"name":"Collect","id":1,"multipath":{"inputs":["A"]}},
        {"name":"egress","id":1}
      ],
      "edges": [
        {"from":"ingress","fromId":1,"to":"A","toId":1},
        {"from":"A","fromId":1,"to":"Collect","toId":1},
        {"from":"Collect","fromId":1,"to":"egress","toId":1}
      ]
    }
  ]
}
```

## How to use it

- Map edges to exact step instances via `fromId`/`toId` when names repeat (e.g., multiple `Collect`).
- Inspect `multipath.merge` to confirm buffer, stability, and sort normalization for `Collect` nodes.
- Check `concurrency` for build/run-time expectations or to flag missing/invalid pragmas.

## Related docs

- `docs/edges.md` — edge kinds, merge semantics, and validation examples.
- `docs/runtime-tests.md` — runtime harness and default ErrorPipeline.

