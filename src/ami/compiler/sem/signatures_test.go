package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestBuildFunctionSignatures_Basics(t *testing.T) {
    f := &ast.File{}
    f.Decls = append(f.Decls, &ast.FuncDecl{
        Name:   "F",
        Params: []ast.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "Event<string>"}},
        Results: []ast.Result{{Type: "bool"}},
    })
    sigs := BuildFunctionSignatures(f)
    s, ok := sigs["F"]
    if !ok { t.Fatalf("signature for F not found") }
    if got := s.String(); got != "func(int,Event<string>) -> (bool)" {
        t.Fatalf("string: %s", got)
    }
}

