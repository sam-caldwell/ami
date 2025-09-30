# Connectivity Cookbook

Quick patterns to diagnose and fix pipeline graph connectivity issues.

## Unreachable From Ingress
- Symptom: `E_PIPELINE_UNREACHABLE_FROM_INGRESS` or StageB `W_PIPELINE_UNREACHABLE_NODE`.
- Cause: Node has no path from `ingress`.
- Example (error):
```
pipeline P(){
  ingress; A; B; egress;
  A -> egress;
  B -> egress;   // B never receives input from ingress
}
```
- Fix: connect upstream of B from a reachable node:
```
  A -> B; B -> egress;
```

## Cannot Reach Egress
- Symptom: `E_PIPELINE_CANNOT_REACH_EGRESS` or StageB `W_PIPELINE_NONTERMINATING_NODE`.
- Cause: Node has no path to `egress`.
- Example (error):
```
pipeline P(){
  ingress; A; B; egress;
  ingress -> A; A -> B;   // B has no outbound edge
}
```
- Fix: add an edge to `egress` (or onward):
```
  B -> egress;
```

## Disconnected Node
- Symptom: `E_PIPELINE_NODE_DISCONNECTED` or StageB `W_PIPELINE_DISCONNECTED_NODE`.
- Cause: Node has degree 0 (no incident edges).
- Example (error):
```
pipeline P(){
  ingress; A; X; egress;
  ingress -> A; A -> egress;
  // X is orphaned
}
```
- Fix: wire X or remove it:
```
  ingress -> X; X -> egress;
```

## Edge Direction Errors
- To ingress forbidden: `E_EDGE_TO_INGRESS`.
- From egress forbidden: `E_EDGE_FROM_EGRESS`.
- Example (error):
```
  A -> ingress;      // invalid
  egress -> A;       // invalid
```
- Fix: edges flow from `ingress` to `egress` only.

## Edges Without Ingress/Egress
- Symptom: `E_EDGES_WITHOUT_INGRESS` / `E_EDGES_WITHOUT_EGRESS` when any edges exist.
- Fix: ensure both sentinel nodes appear when you declare edges:
```
pipeline P(){ ingress; ...; egress; ...edges... }
```

