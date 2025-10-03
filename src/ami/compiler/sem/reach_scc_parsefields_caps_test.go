package sem

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func Test_parseStructFieldsText(t *testing.T) {
    if m, ok := parseStructFieldsText("Struct{a:int, b:bool}"); !ok || len(m) != 2 { t.Fatalf("parse: %v %v", m, ok) }
    if m, ok := parseStructFieldsText("Struct{}"); !ok || len(m) != 0 { t.Fatalf("empty: %v %v", m, ok) }
    if _, ok := parseStructFieldsText("Struct{a int}"); ok { t.Fatalf("expected fail") }
}

func Test_ReachableFunctions(t *testing.T) {
    f := &ast.File{}
    f.Decls = append(f.Decls, &ast.FuncDecl{Name: "main", Body: &ast.BlockStmt{}})
    f.Decls = append(f.Decls, &ast.FuncDecl{Name: "foo", Body: &ast.BlockStmt{Stmts: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Name: "foo"}}}}})
    got := ReachableFunctions(f)
    if !(got["main"] && got["foo"]) { t.Fatalf("reach: %v", got) }
}

func Test_ComputeSCC_Simple(t *testing.T) {
    f := &ast.File{}
    f.Decls = append(f.Decls, &ast.FuncDecl{Name: "a", Body: &ast.BlockStmt{Stmts: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Name: "b"}}}}})
    f.Decls = append(f.Decls, &ast.FuncDecl{Name: "b", Body: &ast.BlockStmt{Stmts: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Name: "a"}}}}})
    scc := ComputeSCC(f)
    if len(scc) != 2 { t.Fatalf("scc: %v", scc) }
}

func Test_AnalyzeCapabilities_TrustAndCapability(t *testing.T) {
    // Missing capability → error
    f := &ast.File{Pragmas: []ast.Pragma{{Domain:"capabilities", Args: []string{"net"}}}}
    f.Decls = append(f.Decls, &ast.PipelineDecl{Name: "P", Stmts: []ast.Stmt{&ast.StepStmt{Name: "io.read"}}})
    ds := AnalyzeCapabilities(f)
    if len(ds) == 0 { t.Fatalf("expected capability error") }
    // Untrusted trust level forbids io.*
    f2 := &ast.File{Pragmas: []ast.Pragma{{Domain:"trust", Params: map[string]string{"level":"untrusted"}}}}
    f2.Decls = append(f2.Decls, &ast.PipelineDecl{Name: "P", Stmts: []ast.Stmt{&ast.StepStmt{Name: "io.read"}}})
    ds2 := AnalyzeCapabilities(f2)
    if len(ds2) == 0 { t.Fatalf("expected trust violation") }
}

