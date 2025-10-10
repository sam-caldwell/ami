package driver

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
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

// Verify literal body lowering for primitive returns and ev identity.
func TestLowerInlineWorkers_LiteralBody_Returns(t *testing.T) {
    // Case 1: primitive literal return
    f := &ast.File{PackageName: "app"}
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (int, error) { return 2 + 3, nil }"}}}
    pd.Stmts = append(pd.Stmts, st)
    f.Decls = append(f.Decls, pd)
    fns := lowerInlineWorkers("app", "u", f)
    if len(fns) != 1 { t.Fatalf("expected 1 fn, got %d", len(fns)) }
    if fns[0].Results[0].Type != "int" { t.Fatalf("want int result, got %+v", fns[0].Results) }
    // Expect three instructions: err null, lit c0, lit c1, add, return (some scaffolds may differ)
    if len(fns[0].Blocks) == 0 || len(fns[0].Blocks[0].Instr) < 3 { t.Fatalf("expected instructions in block: %+v", fns[0].Blocks) }

    // Case 2: identity Event return
    f2 := &ast.File{PackageName: "app"}
    pd2 := &ast.PipelineDecl{Name: "Q"}
    st2 := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (Event<int>, error) { return ev }"}}}
    pd2.Stmts = append(pd2.Stmts, st2)
    f2.Decls = append(f2.Decls, pd2)
    g := lowerInlineWorkers("app", "u", f2)
    if len(g) != 1 { t.Fatalf("expected 1 fn, got %d", len(g)) }
    if g[0].Results[0].Type != "Event<int>" { t.Fatalf("want Event<int>, got %+v", g[0].Results) }
}

// Verify comparisons of numeric literals are lowered to bool results.
func TestLowerInlineWorkers_LiteralBody_Comparisons(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    pd := &ast.PipelineDecl{Name: "P"}
    // explicit bool result
    st := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (bool, error) { return 2 >= 3, nil }"}}}
    pd.Stmts = append(pd.Stmts, st)
    f.Decls = append(f.Decls, pd)
    fns := lowerInlineWorkers("app", "u", f)
    if len(fns) != 1 { t.Fatalf("expected 1 fn, got %d", len(fns)) }
    if fns[0].Results[0].Type != "bool" { t.Fatalf("want bool, got %+v", fns[0].Results) }
    if len(fns[0].Blocks) == 0 || len(fns[0].Blocks[0].Instr) < 3 { t.Fatalf("expected comparison instrs, got %+v", fns[0].Blocks) }
}

// Verify if/else lowering for literal condition and literal/ev returns.
func TestLowerInlineWorkers_IfElse_Returns(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    pd := &ast.PipelineDecl{Name: "P"}
    // if 1 < 2 { return 5 } else { return 7 }
    st := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (int, error) { if 1 < 2 { return 5 } else { return 7 } }"}}}
    pd.Stmts = append(pd.Stmts, st)
    f.Decls = append(f.Decls, pd)
    fns := lowerInlineWorkers("app", "u", f)
    if len(fns) != 1 { t.Fatalf("expected 1 fn, got %d", len(fns)) }
    if len(fns[0].Blocks) != 3 { t.Fatalf("expected 3 blocks (entry, then, else), got %d", len(fns[0].Blocks)) }
}

// Verify event-driven arithmetic: ev + 3 lowers payload extraction + add.
func TestLowerInlineWorkers_EventPayload_Arithmetic(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (int, error) { return ev + 3 }"}}}
    pd.Stmts = append(pd.Stmts, st)
    f.Decls = append(f.Decls, pd)
    fns := lowerInlineWorkers("app", "u", f)
    if len(fns) != 1 { t.Fatalf("expected 1 fn, got %d", len(fns)) }
    if len(fns[0].Blocks) == 0 || len(fns[0].Blocks[0].Instr) < 3 { t.Fatalf("expected payload extract + add, got %+v", fns[0].Blocks) }
}

// Verify unary negation and not
func TestLowerInlineWorkers_Unary_Forms(t *testing.T) {
    // neg literal
    f1 := &ast.File{PackageName: "app"}
    pd1 := &ast.PipelineDecl{Name: "P"}
    st1 := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (int, error) { return -3 }"}}}
    pd1.Stmts = append(pd1.Stmts, st1)
    f1.Decls = append(f1.Decls, pd1)
    g1 := lowerInlineWorkers("app", "u", f1)
    if len(g1) != 1 { t.Fatalf("neg literal: expected 1 fn") }

    // not literal
    f2 := &ast.File{PackageName: "app"}
    pd2 := &ast.PipelineDecl{Name: "Q"}
    st2 := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (bool, error) { return !true }"}}}
    pd2.Stmts = append(pd2.Stmts, st2)
    f2.Decls = append(f2.Decls, pd2)
    g2 := lowerInlineWorkers("app", "u", f2)
    if len(g2) != 1 { t.Fatalf("not literal: expected 1 fn") }
}

// Verify tiny let/assignment patterns are lowered equivalently to direct returns.
func TestLowerInlineWorkers_Let_Assign_Return(t *testing.T) {
    // let x = 5; return x
    f := &ast.File{PackageName: "app"}
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (int, error) { let x = 5; return x }"}}}
    pd.Stmts = append(pd.Stmts, st)
    f.Decls = append(f.Decls, pd)
    fns := lowerInlineWorkers("app", "u", f)
    if len(fns) != 1 { t.Fatalf("expected 1 fn, got %d", len(fns)) }
    if fns[0].Results[0].Type != "int" { t.Fatalf("want int result, got %+v", fns[0].Results) }
    if len(fns[0].Blocks) == 0 || len(fns[0].Blocks[0].Instr) < 2 { t.Fatalf("expected instructions for let, got %+v", fns[0].Blocks) }

    // x := ev; return x
    f2 := &ast.File{PackageName: "app"}
    pd2 := &ast.PipelineDecl{Name: "Q"}
    st2 := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (int, error) { x := ev; return x }"}}}
    pd2.Stmts = append(pd2.Stmts, st2)
    f2.Decls = append(f2.Decls, pd2)
    g := lowerInlineWorkers("app", "u", f2)
    if len(g) != 1 { t.Fatalf("expected 1 fn, got %d", len(g)) }
    if len(g[0].Blocks) == 0 || len(g[0].Blocks[0].Instr) < 2 { t.Fatalf("expected payload extract for ev assign, got %+v", g[0].Blocks) }
}

// Verify payload field return lowers to field helper op
func TestLowerInlineWorkers_PayloadField_Return(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Transform", Args: []ast.Arg{{Text: "worker=func(ev Event<int>) (int, error) { return event.payload.field.k }"}}}
    pd.Stmts = append(pd.Stmts, st)
    f.Decls = append(f.Decls, pd)
    fns := lowerInlineWorkers("app", "u", f)
    if len(fns) != 1 { t.Fatalf("expected 1 fn, got %d", len(fns)) }
    if len(fns[0].Blocks) == 0 { t.Fatalf("no blocks") }
    found := false
    for _, ins := range fns[0].Blocks[0].Instr {
        if e, ok := ins.(ir.Expr); ok && len(e.Op) > 0 && len(e.Args) == 1 && e.Op == "event.payload.field.k" {
            found = true
            break
        }
    }
    if !found { t.Fatalf("missing event.payload.field.k op in lowered IR: %+v", fns[0].Blocks[0].Instr) }
}
