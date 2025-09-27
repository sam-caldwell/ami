package types

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestFromAST_PrimitivesAndGenerics(t *testing.T) {
    if FromAST("int").String() != "int" { t.Fatalf("int") }
    if FromAST("string").String() != "string" { t.Fatalf("string") }
    if FromAST("Event<int>").String() != "Event<int>" { t.Fatalf("Event<int>") }
    if FromAST("Error<string>").String() != "Error<string>" { t.Fatalf("Error<string>") }
    if FromAST("Owned<T>").String() != "Owned<T>" { t.Fatalf("Owned<T>") }
}

func TestFromAST_ContainersAndPointers(t *testing.T) {
    if FromAST("map<string,int64>").String() != "map<string,int64>" { t.Fatalf("map") }
    if FromAST("slice<int>").String() != "slice<int>" { t.Fatalf("slice<T>") }
    if FromAST("set<T>").String() != "set<T>" { t.Fatalf("set<T>") }
    if FromAST("[]string").String() != "[]string" { t.Fatalf("[]string") }
    if FromAST("*int").String() != "*int" { t.Fatalf("*int") }
}

func TestBuildFunction_FromFuncDecl(t *testing.T) {
    fn := &ast.FuncDecl{
        Params:  []ast.Param{{Name: "a", Type: "int"}, {Name: "b", Type: "Event<T>"}},
        Results: []ast.Result{{Type: "map<string,int64>"}},
    }
    ft := BuildFunction(fn)
    got := ft.String()
    want := "func(int,Event<T>) -> (map<string,int64>)"
    if got != want { t.Fatalf("function string: %s != %s", got, want) }
}
