package ir

import (
    "encoding/json"
    "testing"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestLowerPipelines_InlineWorker_IncludedWithTypes(t *testing.T) {
    src := `package p
pipeline P { Ingress(cfg).Transform(worker=func(ev Event<string>) Event<string> { return ev }).Egress() }`
    p := parser.New(src)
    f := p.ParseFile()
    if len(p.Errors()) > 0 { t.Fatalf("parser errors: %+v", p.Errors()) }
    m := &Module{Package: "p", Unit: "u.ami"}
    m.LowerPipelines(f)
    sch := m.ToPipelinesSchema()
    // Serialize to ensure new fields marshal
    if _, err := json.Marshal(sch); err != nil { t.Fatalf("marshal pipelines: %v", err) }
    if len(sch.Pipelines) != 1 || len(sch.Pipelines[0].Steps) < 2 {
        t.Fatalf("unexpected pipelines shape: %+v", sch)
    }
    // Transform is step1
    st := sch.Pipelines[0].Steps[1]
    var found bool
    for _, w := range st.Workers {
        if w.Kind == "inline" && w.Origin == "literal" && w.Input == "string" && w.OutputKind == "Event" && w.Output == "string" {
            found = true
            break
        }
    }
    if !found { t.Fatalf("inline worker lowering not found in workers: %#v", st.Workers) }
    // Attrs should include worker key
    if st.Attrs == nil || st.Attrs["worker"] == "" {
        t.Fatalf("expected step attrs to include worker; got %+v", st.Attrs)
    }
    _ = astpkg.TypeRef{} // silence unused in older Go if imports reordered
}
