# graph.v1 — Pipeline Graph Schema (Visualize)

Fields

- `schema`: fixed `graph.v1`.
- `nodes`: array of `{ id, label }`.
- `edges`: array of `{ from, to }`.
- Optional: per-node attributes where available (scaffold level).

Example

```
{
  "schema": "graph.v1",
  "nodes": [
    {"id":"ingress","label":"Ingress"},
    {"id":"work","label":"Transform"},
    {"id":"egress","label":"Egress"}
  ],
  "edges": [
    {"from":"ingress","to":"work"},
    {"from":"work","to":"egress"}
  ]
}
```
