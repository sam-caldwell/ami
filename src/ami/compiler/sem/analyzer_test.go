package sem

import (
    "testing"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestAnalyzeFile_DuplicateFunc(t *testing.T) {
    f := &astpkg.File{Decls: []astpkg.Node{
        astpkg.FuncDecl{Name: "main"},
        astpkg.FuncDecl{Name: "helper"},
        astpkg.FuncDecl{Name: "helper"}, // duplicate
    }}
    res := AnalyzeFile(f)
    if len(res.Diagnostics) != 1 { t.Fatalf("want 1 diag; got %d", len(res.Diagnostics)) }
    if res.Diagnostics[0].Code != "E_DUP_FUNC" { t.Fatalf("diag code=%s", res.Diagnostics[0].Code) }
    if res.Scope.Lookup("main") == nil || res.Scope.Lookup("helper") == nil {
        t.Fatalf("expected functions in scope")
    }
}

