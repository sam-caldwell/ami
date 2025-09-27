package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestDebug_ReturnsCount_TwoReturns(t *testing.T) {
    src := "package app\nfunc H(){ return slice<int>{1}; return slice<string>{\"x\"} }"
    f := &source.File{Name: "ri_dbg.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    count := 0
    for _, d := range af.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil {
            for _, st := range fn.Body.Stmts {
                if _, ok := st.(*ast.ReturnStmt); ok { count++ }
            }
        }
    }
    if count != 2 { t.Fatalf("expected 2 returns, got %d", count) }
}
