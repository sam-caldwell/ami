package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeAmbiguity_EmptySliceLiteral_EmitsDiag(t *testing.T) {
    // func F() { var x = slice<any>{} }
    file := &ast.File{}
    body := &ast.BlockStmt{}
    body.Stmts = append(body.Stmts, &ast.VarDecl{Init: &ast.SliceLit{Pos: source.Position{Line: 1, Column: 1}, TypeName: "", Elems: nil}})
    fd := &ast.FuncDecl{Name: "F", Body: body}
    file.Decls = append(file.Decls, fd)
    diags := AnalyzeAmbiguity(file)
    if len(diags) == 0 { t.Fatalf("expected at least one ambiguity diagnostic") }
    if diags[0].Code != "E_TYPE_AMBIGUOUS" { t.Fatalf("unexpected code: %s", diags[0].Code) }
}

