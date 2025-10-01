# Runtime Test Harness

Overview of `ami test` runtime execution support, KV store integration, and error pipeline emission.

- CLI flags: `--kv-metrics`, `--kv-dump`, `--kv-events`, per-case `emit=true`.
- Artifacts: process-level under `build/test/kv/` and per-case under `build/test/kv/<file>_<case>.*.json`.
- Default ErrorPipeline on error cases emits `errors.v1` JSON lines to stderr; toggles: `--no-errorpipe`, `--errorpipe-human`.

See also:
- `docs/toolchain/runtime-kvstore.md` — KV store design and API
- `docs/toolchain/pipelines-v1-quickstart.md` — debug artifact navigation

## Exec Test Helpers (for Contributors)

To write focused runtime executor tests without full compiler context, use the helpers under `src/ami/runtime/exec` (test-only):

- `WriteEdges(t, pkg, pipeline, edges)`: write `build/debug/asm/<pkg>/edges.json` with a minimal `edges.v1` index.
- `MakeModuleWithEdges(t, pkg, pipeline, edges)`: write edges and return a minimal `ir.Module` with that pipeline.
- `AttachCollect(m, pipeline, step, plan)`: append a Collect spec (step name + `ir.MergePlan`) to a module.
- `MakeCollectOnlyModule(t, pkg, pipeline, step, plan)`: `ingress -> step -> egress` plus attached plan.
- `MakeTransformOnlyModule(t, pkg, pipeline, transforms)`: build edges `ingress -> T1 -> T2 -> ... -> egress`.
- `MakeTransformAndCollectModule(t, pkg, pipeline, transforms, collectStep, plan)`: transforms followed by one Collect.
- `NewMergePlan()` fluent builder with common knobs:
  - `.Stable(bool)`, `.Key(field)`, `.PartitionBy(field)`, `.Buffer(cap, policy)`, `.Window(n)`, `.TimeoutMs(ms)`,
    `.SortAsc(field)`, `.SortDesc(field)`, `.Watermark(field, latenessMs)`, `.Dedup(field)`, `.LatePolicy(policy)`, `.Build()`.

Example: Timer -> Transform -> Collect

```go
edges := []edgeEntry{{From:"ingress", To:"Timer"}, {From:"Timer", To:"X"}, {From:"X", To:"Collect"}, {From:"Collect", To:"egress"}}
m := MakeModuleWithEdges(t, "app", "P", edges)
m = AttachCollect(m, "P", "Collect", NewMergePlan().SortAsc("ts").Window(1).Build())
eng, _ := NewEngineFromModule(m)
out, stats, _ := eng.RunPipelineWithStats(ctx, m, "P", in, nil, "none", "add_field:flag", ExecOptions{TimerInterval: 5*time.Millisecond, TimerCount: 3, Sandbox: SandboxPolicy{AllowDevice:true}})
for e := range out { /* assert payload */ }
for range stats { /* drain */ }
```

Notes:
- Edges must include both `ingress` and `egress`. Use step names consistently across edges and Collect specs.
- Timer ingress is activated by including a `Timer` node in edges. Device capability must be allowed in `SandboxPolicy`.
