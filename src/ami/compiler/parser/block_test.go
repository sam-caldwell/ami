package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParseBlock_Internal exercises the private parseBlock depth logic.
func TestParseBlock_Internal(t *testing.T) {
    code := "{{}}"
    f := (&source.FileSet{}).AddFile("b.ami", code)
    p := New(f)
    if p.cur.Lexeme != "{" { t.Fatalf("want '{', got %q", p.cur.Lexeme) }
    if _, err := p.parseBlock(); err != nil { t.Fatalf("parseBlock: %v", err) }
}

// TestParseFuncBlock_Statements covers var/assign/defer/return and comment attachment.
func TestParseFuncBlock_Statements(t *testing.T) {
    src := "package app\nfunc F(){\n// c1\nvar x int = 1;\n*x = 2;\nx = x + 3;\ndefer G();\nreturn x\n}\nfunc G(){}\n"
    f := (&source.FileSet{}).AddFile("f.ami", src)
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("ParseFile: %v", err) }
    if len(file.Decls) < 1 { t.Fatalf("no decls parsed") }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok || fn.Body == nil || len(fn.Body.Stmts) < 4 { t.Fatalf("body not parsed with statements: %#v", file.Decls[0]) }
}
