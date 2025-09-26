package ir

import (
    "encoding/json"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

// Verify normalized merge config appears in pipelines.v1 and edges.v1 snapshots.
func TestPipelinesSchema_MultiPath_NormalizedMerge(t *testing.T) {
    src := `package p
pipeline Up { Ingress(cfg).Transform(w).Egress() }
func w(ev Event<int>) (Event<int>, error) {}
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[ edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=block,type=int), edge.Pipeline(name=Up,minCapacity=0,maxCapacity=0,backpressure=dropNewest,type=int) ], merge=Sort("ts","desc"), merge=Stable(), merge=Buffer(10,dropOldest))).Egress() }
`
    p := parser.New(src)
    f := p.ParseFile()
    m := Module{Package: "p", Unit: "u.ami"}
    m.LowerPipelines(f)
    sch := m.ToPipelinesSchema()
    b, err := json.Marshal(sch)
    if err != nil || len(b) == 0 { t.Fatalf("marshal: %v", err) }
    if len(sch.Pipelines) != 2 { t.Fatalf("want 2 pipelines; got %+v", sch.Pipelines) }
    // Find P.Collect step
    var found bool
    for _, pl := range sch.Pipelines {
        if pl.Name != "P" { continue }
        if len(pl.Steps) < 2 { t.Fatalf("P has too few steps: %d", len(pl.Steps)) }
        st := pl.Steps[1]
        if st.InEdge == nil || st.InEdge.MultiPath == nil || st.InEdge.MultiPath.MergeConfig == nil {
            t.Fatalf("missing mergeConfig in MultiPath: %+v", st.InEdge)
        }
        cfg := st.InEdge.MultiPath.MergeConfig
        if cfg.SortField != "ts" || cfg.SortOrder != "desc" || !cfg.Stable || cfg.BufferCapacity != 10 || cfg.BufferBackpressure != "dropOldest" {
            t.Fatalf("unexpected merge config: %+v", cfg)
        }
        // Golden JSON for just mergeConfig (deterministic subset)
        cb, _ := json.Marshal(cfg)
        want := `{"sortField":"ts","sortOrder":"desc","stable":true,"bufferCapacity":10,"bufferBackpressure":"dropOldest"}`
        if string(cb) != want { t.Fatalf("mergeConfig golden mismatch:\n got: %s\nwant: %s", string(cb), want) }
        found = true
    }
    if !found { t.Fatalf("pipeline P not found or missing merge config") }
}
