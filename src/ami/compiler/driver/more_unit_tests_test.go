package driver

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_joinCSV(t *testing.T) {
    if s := joinCSV(nil); s != "" { t.Fatalf("empty: %q", s) }
    if s := joinCSV([]string{"a"}); s != "a" { t.Fatalf("one: %q", s) }
    if s := joinCSV([]string{"a","b","c"}); s != "a,b,c" { t.Fatalf("many: %q", s) }
}

func Test_lowerBlock_Empty(t *testing.T) {
    st := &lowerState{}
    b := &ast.BlockStmt{}
    got := lowerBlock(st, b)
    if len(got) != 0 { t.Fatalf("expected no instrs, got %d", len(got)) }
}

func Test_lowerStmtAssign_ValueAndFallback(t *testing.T) {
    st := &lowerState{varTypes: map[string]string{}}
    // Assign numeric literal: should propagate type to dest
    as1 := &ast.AssignStmt{Name: "x", Value: &ast.NumberLit{Text: "1"}}
    ins1 := lowerStmtAssign(st, as1)
    if _, ok := ins1.(ir.Assign); !ok || st.varTypes["x"] == "" {
        t.Fatalf("expected assign and type propagation: %+v, types=%v", ins1, st.varTypes)
    }
    // Fallback: expression that fails to lower yields temp of type any
    as2 := &ast.AssignStmt{Name: "y", Value: &ast.DurationLit{Text: "bogus"}}
    ins2 := lowerStmtAssign(st, as2)
    if _, ok := ins2.(ir.Assign); !ok {
        t.Fatalf("expected assign fallback, got %T", ins2)
    }
}

func Test_lowerStmtReturn_CollectsValues(t *testing.T) {
    st := &lowerState{}
    rs := &ast.ReturnStmt{Results: []ast.Expr{&ast.NumberLit{Text: "1"}, &ast.StringLit{Value: "x"}}}
    ins := lowerStmtReturn(st, rs)
    if _, ok := ins.(ir.Return); !ok {
        t.Fatalf("expected RETURN instruction, got %T", ins)
    }
}

func Test_parseDurationMs(t *testing.T) {
    cases := map[string]int{
        "100": 100,
        "200ms": 200,
        "1s": 1000,
        "2m": 120000,
        "1h": 3600000,
        " \"3s\" ": 3000,
    }
    for in, want := range cases {
        if got, ok := parseDurationMs(in); !ok || got != want {
            t.Fatalf("%s => %d,%v (want %d,true)", in, got, ok, want)
        }
    }
    if _, ok := parseDurationMs("bad"); ok {
        t.Fatalf("expected bad parse to fail")
    }
}

func Test_handlerTokenImmediate(t *testing.T) {
    // Ident
    if tok, ok := handlerTokenImmediate(&ast.IdentExpr{Name: "MyHandler"}); !ok || tok == "" {
        t.Fatalf("ident: got %q,%v", tok, ok)
    }
    // Selector
    if tok, ok := handlerTokenImmediate(&ast.SelectorExpr{X: &ast.IdentExpr{Name: "pkg"}, Sel: "Fn"}); !ok || tok == "" {
        t.Fatalf("selector: got %q,%v", tok, ok)
    }
    // Expr offset fallback
    if tok, ok := handlerTokenImmediate(&ast.StringLit{Pos: source.Position{Offset: 42}, Value: "x"}); !ok || tok == "" {
        t.Fatalf("offset: got %q,%v", tok, ok)
    }
}
