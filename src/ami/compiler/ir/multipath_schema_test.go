package ir

import (
    "encoding/json"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestPipelinesSchema_MultiPath_Scaffold(t *testing.T) {
    src := `package p
pipeline P {
  Ingress(cfg).Collect(in=edge.MultiPath(inputs=[ edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=block,type=int), edge.Pipeline(name=Up, minCapacity=0,maxCapacity=0,backpressure=dropNewest,type=int) ], merge=Sort("event.ts","asc"))).Egress()
}`
    p := parser.New(src)
    f := p.ParseFile()
    m := Module{Package: "p", Unit: "u.ami"}
    m.LowerPipelines(f)
    sch := m.ToPipelinesSchema()
    // Marshal to ensure schema writes
    if _, err := json.Marshal(sch); err != nil { t.Fatalf("marshal: %v", err) }
    if len(sch.Pipelines) != 1 || len(sch.Pipelines[0].Steps) < 2 {
        t.Fatalf("unexpected pipelines shape: %+v", sch)
    }
    // Step 1 is Collect in this chain
    st := sch.Pipelines[0].Steps[1]
    if st.InEdge == nil || st.InEdge.Kind != "edge.MultiPath" || st.InEdge.MultiPath == nil {
        t.Fatalf("expected MultiPath edge; got %+v", st.InEdge)
    }
    if len(st.InEdge.MultiPath.Inputs) != 2 {
        t.Fatalf("expected 2 inputs; got %d", len(st.InEdge.MultiPath.Inputs))
    }
    if len(st.InEdge.MultiPath.Merge) == 0 || st.InEdge.MultiPath.Merge[0].Name == "" {
        t.Fatalf("expected merge op; got %+v", st.InEdge.MultiPath.Merge)
    }
}

