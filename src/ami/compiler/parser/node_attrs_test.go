package parser

import (
    "strings"
    "testing"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func findPipeline(t *testing.T, f *astpkg.File) astpkg.PipelineDecl {
    t.Helper()
    for _, d := range f.Decls {
        if p, ok := d.(astpkg.PipelineDecl); ok {
            return p
        }
    }
    return astpkg.PipelineDecl{}
}

func TestParser_NodeAttributes_Parsed(t *testing.T) {
    src := `package p
pipeline P {
  Ingress(in=edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=block,type=[]byte)).
  Transform(worker=doThing,minWorkers=1,maxWorkers=2,onError=drop,capabilities=net).
  Egress()
}`
    p := New(src)
    f := p.ParseFile()
    pipe := findPipeline(t, f)
    if pipe.Name != "P" || len(pipe.Steps) < 2 {
        t.Fatalf("expected pipeline P with steps; got %+v", pipe)
    }
    // Ingress attrs
    ing := pipe.Steps[0]
    if got := strings.TrimSpace(ing.Attrs["in"]); !strings.HasPrefix(got, "edge.FIFO(") {
        t.Fatalf("expected ingress in=edge.FIFO(...), got %q", got)
    }
    if ing.Attrs["type"] == "" { // inner type key inside FIFO should be preserved in arg string; top-level type attr optional
        // OK: type is inside FIFO, not at top-level; ensure arg retained
        if len(ing.Args) == 0 || !strings.Contains(ing.Args[0], "type=[]byte") {
            t.Fatalf("expected raw args to retain inner attributes; got %v", ing.Args)
        }
    }
    // Transform attrs
    tr := pipe.Steps[1]
    if tr.Attrs["worker"] != "doThing" {
        t.Fatalf("expected worker=doThing, got %q", tr.Attrs["worker"])
    }
    if tr.Attrs["minWorkers"] != "1" || tr.Attrs["maxWorkers"] != "2" {
        t.Fatalf("expected min/max workers, got %+v", tr.Attrs)
    }
    if tr.Attrs["onError"] != "drop" || tr.Attrs["capabilities"] != "net" {
        t.Fatalf("expected onError/drop and capabilities/net, got %+v", tr.Attrs)
    }
    // Ensure workers slice captures worker attribute too
    if len(tr.Workers) == 0 || tr.Workers[0].Name != "doThing" {
        t.Fatalf("expected Workers to include doThing, got %+v", tr.Workers)
    }
}

