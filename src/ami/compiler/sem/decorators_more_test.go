package sem

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func Test_AnalyzeDecorators_WorkerAndUndefined(t *testing.T) {
    // Worker signature with decorator → error
    fn1 := &ast.FuncDecl{Name: "Worker", Params: []ast.Param{{Type: "Event"}}, Results: []ast.Result{{Type: "Event"}, {Type: "error"}}, Decorators: []ast.Decorator{{Name: "metrics"}}}
    // Non-worker with undefined decorator
    fn2 := &ast.FuncDecl{Name: "F", Params: []ast.Param{{Type: "int"}}, Results: []ast.Result{{Type: "int"}}, Decorators: []ast.Decorator{{Name: "unknown.decorator"}}}
    f := &ast.File{Decls: []ast.Decl{fn1, fn2}}
    ds := AnalyzeDecorators(f)
    if len(ds) == 0 { t.Fatalf("expected decorator diagnostics") }
}

