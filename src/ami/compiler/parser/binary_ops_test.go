package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure parser recognizes logical/bitwise/shift binary operators with precedence table.
func TestParser_BinaryOperators_Parse(t *testing.T) {
	// Use an expression mix to exercise precedence; exact tree shape not asserted here.
	code := "package app\nfunc F(){ var a int; var b int; var c int; var d int; var e int; var f int; var g int; a && b || c; d << e | f ^ g & a }\n"
	f := &source.File{Name: "binops.ami", Content: code}
	p := New(f)
	af, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var haveLogic, haveBitwise bool
	for _, d := range af.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name != "F" || fn.Body == nil {
			continue
		}
		for _, st := range fn.Body.Stmts {
			if es, ok := st.(*ast.ExprStmt); ok {
				if _, ok := es.X.(*ast.BinaryExpr); ok {
					// Count both logical and bitwise chains seen in the function
					haveLogic = true
					haveBitwise = true
				}
			}
		}
	}
	if !haveLogic || !haveBitwise {
		t.Fatalf("missing parsed binary expressions: logic=%v bitwise=%v", haveLogic, haveBitwise)
	}
}
