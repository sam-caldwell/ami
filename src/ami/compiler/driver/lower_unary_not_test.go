package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestLowerExpr_UnaryNot(t *testing.T) {
    // Build a minimal state with a boolean variable 'x'
    st := &lowerState{varTypes: map[string]string{"x": "bool"}}
    u := &ast.UnaryExpr{Op: token.Bang, X: &ast.IdentExpr{Name: "x"}}
    ex, ok := lowerExpr(st, u)
    if !ok { t.Fatalf("lowering failed") }
    if ex.Op != "not" { t.Fatalf("expected op 'not', got %q", ex.Op) }
}
