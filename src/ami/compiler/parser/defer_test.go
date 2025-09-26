package parser

import (
	ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestParser_Defer_ParsesCall(t *testing.T) {
	src := `package p
func f(ev Event<string>) (Event<string>, error) {
    defer cleanup()
}`
	p := New(src)
	f := p.ParseFile()
	if len(f.Decls) != 1 {
		t.Fatalf("expected 1 decl; got %d", len(f.Decls))
	}
	fd, ok := f.Decls[0].(ast.FuncDecl)
	if !ok {
		t.Fatalf("decl type = %T; want FuncDecl", f.Decls[0])
	}
	// BodyStmts should contain a DeferStmt(CallExpr)
	if len(fd.BodyStmts) != 1 {
		t.Fatalf("expected 1 body stmt; got %d", len(fd.BodyStmts))
	}
	ds, ok := fd.BodyStmts[0].(ast.DeferStmt)
	if !ok {
		t.Fatalf("stmt type = %T; want DeferStmt", fd.BodyStmts[0])
	}
	call, ok := ds.X.(ast.CallExpr)
	if !ok {
		t.Fatalf("defer X type = %T; want CallExpr", ds.X)
	}
	if id, ok := call.Fun.(ast.Ident); !ok || id.Name != "cleanup" {
		t.Fatalf("callee = %#v; want Ident('cleanup')", call.Fun)
	}
}

func TestParser_Defer_MethodCall(t *testing.T) {
	src := `package p
func f(ev Event<string>) (Event<string>, error) {
    defer r.Close()
}`
	p := New(src)
	f := p.ParseFile()
	fd := f.Decls[0].(ast.FuncDecl)
	if len(fd.BodyStmts) != 1 {
		t.Fatalf("expected 1 stmt; got %d", len(fd.BodyStmts))
	}
	ds, ok := fd.BodyStmts[0].(ast.DeferStmt)
	if !ok {
		t.Fatalf("stmt type = %T; want DeferStmt", fd.BodyStmts[0])
	}
	call, ok := ds.X.(ast.CallExpr)
	if !ok {
		t.Fatalf("defer X type = %T; want CallExpr", ds.X)
	}
	if sel, ok := call.Fun.(ast.SelectorExpr); !ok || sel.Sel != "Close" {
		t.Fatalf("callee = %#v; want SelectorExpr(..., 'Close')", call.Fun)
	}
}
