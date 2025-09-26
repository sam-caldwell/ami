package parser

import (
    "strings"
    "testing"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Ensure parser round-trips MultiPath attribute on Collect nodes as a raw attr string.
func TestParser_NodeAttrs_MultiPath_OnCollect(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Collect(in=edge.MultiPath(inputs=[edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=block,type=int)], merge=Sort("ts","asc"))).Egress() }`
    p := New(src)
    f := p.ParseFile()
    // locate pipeline and Collect step
    var found bool
    for _, d := range f.Decls {
        if pd, ok := d.(astpkg.PipelineDecl); ok {
            for _, st := range pd.Steps {
                if strings.EqualFold(st.Name, "collect") {
                    v := strings.TrimSpace(st.Attrs["in"])
                    if !strings.HasPrefix(v, "edge.MultiPath(") {
                        t.Fatalf("expected in=edge.MultiPath(...); got %q", v)
                    }
                    found = true
                }
            }
        }
    }
    if !found { t.Fatalf("Collect step with MultiPath not found") }
}
