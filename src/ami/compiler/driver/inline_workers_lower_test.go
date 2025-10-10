package driver

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Ensure inline func-literal workers synthesize IR functions with deterministic names.
func TestLowerInlineWorkers_SynthesizesFunctions(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    pd := &ast.PipelineDecl{Name: "P"}
    // Transform with inline worker literal
    st := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (Event<int>, error)"}}}
    pd.Stmts = append(pd.Stmts, st)
    f.Decls = append(f.Decls, pd)
    fns := lowerInlineWorkers("app", "u", f)
    if len(fns) != 1 { t.Fatalf("expected 1 generated worker, got %d", len(fns)) }
    if fns[0].Name != "InlineWorker_u_P_1" { t.Fatalf("unexpected name: %s", fns[0].Name) }
    if len(fns[0].Params) != 1 || fns[0].Params[0].Type != "Event<int>" { t.Fatalf("param: %+v", fns[0].Params) }
    if len(fns[0].Results) != 2 || fns[0].Results[0].Type != "Event<int>" || fns[0].Results[1].Type != "error" {
        t.Fatalf("results: %+v", fns[0].Results)
    }
}

