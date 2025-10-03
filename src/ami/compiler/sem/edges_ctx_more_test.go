package sem

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func Test_AnalyzeEdgesInContext_FIFO_ParamsValidation(t *testing.T) {
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{{Name: "edge.FIFO", Args: []ast.Arg{{Text: "min=-1"}, {Text: "unknown=1"}, {Text: "backpressure=bad"}}}}}
    pd.Stmts = append(pd.Stmts, st)
    f := &ast.File{Decls: []ast.Decl{pd}}
    ds := AnalyzeEdgesInContext(f, map[string]string{"Q":"Struct{}"})
    if len(ds) < 2 { t.Fatalf("expected diagnostics, got %v", ds) }
}

func Test_AnalyzeEdgesInContext_FIFO_OrderAndLegacyDrop(t *testing.T) {
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{{Name: "edge.FIFO", Args: []ast.Arg{{Text: "min=10"}, {Text: "max=5"}, {Text: "backpressure=drop"}}}}}
    pd.Stmts = append(pd.Stmts, st)
    f := &ast.File{Decls: []ast.Decl{pd}}
    ds := AnalyzeEdgesInContext(f, map[string]string{"Q":"Struct{}"})
    if len(ds) < 2 { t.Fatalf("expected order and legacy warnings, got %v", ds) }
}

func Test_AnalyzeEdgesInContext_Pipeline_TypeMismatch(t *testing.T) {
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{{Name: "edge.Pipeline", Args: []ast.Arg{{Text: "name=Target"}, {Text: "type=Struct{a:int}"}}}}}
    pd.Stmts = append(pd.Stmts, st)
    f := &ast.File{Decls: []ast.Decl{pd}}
    ctx := map[string]string{"Target": "Struct{b:string}"}
    ds := AnalyzeEdgesInContext(f, ctx)
    // should include type mismatch diagnostic
    if len(ds) == 0 { t.Fatalf("expected type mismatch diagnostic") }
}

func Test_AnalyzeEdgesInContext_Pipeline_NameRequired(t *testing.T) {
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{{Name: "edge.Pipeline", Args: []ast.Arg{{Text: "type=Struct{}"}}}}}
    pd.Stmts = append(pd.Stmts, st)
    f := &ast.File{Decls: []ast.Decl{pd}}
    ds := AnalyzeEdgesInContext(f, map[string]string{"X":"Struct{}"})
    if len(ds) == 0 { t.Fatalf("expected name required diagnostic") }
}
