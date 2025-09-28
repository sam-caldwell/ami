package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestLowerExpr_Bitwise_Shifts_And_Xor(t *testing.T) {
    st := &lowerState{varTypes: map[string]string{"x": "int", "y": "int"}}
    // x << y
    be := &ast.BinaryExpr{Op: token.Shl, X: &ast.IdentExpr{Name: "x"}, Y: &ast.IdentExpr{Name: "y"}}
    ex, ok := lowerExpr(st, be)
    if !ok || ex.Op != "shl" { t.Fatalf("shl: %+v", ex) }
    // x >> y
    be = &ast.BinaryExpr{Op: token.Shr, X: &ast.IdentExpr{Name: "x"}, Y: &ast.IdentExpr{Name: "y"}}
    ex, ok = lowerExpr(st, be)
    if !ok || ex.Op != "shr" { t.Fatalf("shr: %+v", ex) }
    // x ^ y
    be = &ast.BinaryExpr{Op: token.BitXor, X: &ast.IdentExpr{Name: "x"}, Y: &ast.IdentExpr{Name: "y"}}
    ex, ok = lowerExpr(st, be)
    if !ok || ex.Op != "xor" { t.Fatalf("xor: %+v", ex) }
    // x & y
    be = &ast.BinaryExpr{Op: token.BitAnd, X: &ast.IdentExpr{Name: "x"}, Y: &ast.IdentExpr{Name: "y"}}
    ex, ok = lowerExpr(st, be)
    if !ok || ex.Op != "band" { t.Fatalf("band(bit): %+v", ex) }
}
