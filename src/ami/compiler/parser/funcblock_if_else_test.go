package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParseFuncBlock_IfElse(t *testing.T) {
    src := "package app\nfunc F(){ if (1) { return } else { defer G() } }\nfunc G(){}\n"
    f := (&source.FileSet{}).AddFile("ie.ami", src)
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("ParseFile: %v", err) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok || fn.Body == nil { t.Fatalf("no func/body") }
    // ensure if statement parsed with else
    var foundIf bool
    for _, st := range fn.Body.Stmts {
        if is, ok := st.(*ast.IfStmt); ok {
            if is.Then != nil && is.Else != nil { foundIf = true }
        }
    }
    if !foundIf { t.Fatalf("if/else not parsed") }
}

