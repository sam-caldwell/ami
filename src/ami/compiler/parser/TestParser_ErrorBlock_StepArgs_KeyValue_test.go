package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure error block step calls accept key=value style args like pipeline steps.
func TestParser_ErrorBlock_StepArgs_KeyValue(t *testing.T) {
    src := "package app\nerror { Alpha(name=Ok, retries=3) }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) == 0 { t.Fatalf("no decls parsed") }
    eb, ok := file.Decls[0].(*ast.ErrorBlock)
    if !ok { t.Fatalf("decl0 not ErrorBlock: %T", file.Decls[0]) }
    if eb.Body == nil || len(eb.Body.Stmts) == 0 { t.Fatalf("error block empty") }
    st, ok := eb.Body.Stmts[0].(*ast.StepStmt)
    if !ok { t.Fatalf("stmt0 not StepStmt: %T", eb.Body.Stmts[0]) }
    if len(st.Args) < 2 { t.Fatalf("want >=2 args, got %d", len(st.Args)) }
    if st.Args[0].Text != "name=Ok" || st.Args[1].Text != "retries=3" {
        t.Fatalf("unexpected arg texts: %v", []string{st.Args[0].Text, st.Args[1].Text})
    }
}

