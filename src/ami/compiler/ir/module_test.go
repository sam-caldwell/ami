package ir

import (
	"encoding/json"
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestFromAST_ToSchema_SortedFunctions(t *testing.T) {
	f := &astpkg.File{Package: "p", Decls: []astpkg.Node{
		astpkg.FuncDecl{Name: "z"},
		astpkg.FuncDecl{Name: "a"},
		astpkg.FuncDecl{Name: "m"},
	}}
	m := FromASTFile("p", "", "unit.ami", f)
	ir := m.ToSchema()
	if len(ir.Functions) != 3 {
		t.Fatalf("functions=%d", len(ir.Functions))
	}
	if ir.Functions[0].Name != "a" || ir.Functions[1].Name != "m" || ir.Functions[2].Name != "z" {
		t.Fatalf("not sorted: %v", []string{ir.Functions[0].Name, ir.Functions[1].Name, ir.Functions[2].Name})
	}
}

func TestIR_ToSchema_GoldenJSON(t *testing.T) {
	f := &astpkg.File{Package: "p", Decls: []astpkg.Node{astpkg.FuncDecl{Name: "a"}}}
	m := FromASTFile("p", "", "u.ami", f)
	ir := m.ToSchema()
	b, _ := json.Marshal(ir)
	want := `{"schema":"ir.v1","timestamp":"","package":"p","file":"u.ami","functions":[{"name":"a","blocks":[{"label":"entry","instrs":null}]}]}`
	if string(b) != want {
		t.Fatalf("golden mismatch:\n got: %s\nwant: %s", string(b), want)
	}
}
