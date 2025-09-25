package types

import (
    "testing"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func tr(name string, ptr, slice bool, args ...astpkg.TypeRef) astpkg.TypeRef {
    return astpkg.TypeRef{Name: name, Ptr: ptr, Slice: slice, Args: args}
}

func TestFromAST_Mapping(t *testing.T) {
    if FromAST(tr("int", false, false)).String() != "int" { t.Fatal("int map") }
    if FromAST(tr("string", false, false)).String() != "string" { t.Fatal("string map") }
    if FromAST(tr("Event", false, false, tr("string", false, false))).String() != "Event<string>" { t.Fatal("event map") }
    if FromAST(tr("Error", false, false, tr("int", false, false))).String() != "Error<int>" { t.Fatal("error map") }
    if FromAST(tr("map", false, false, tr("string", false, false), tr("int", false, false))).String() != "map<string,int>" { t.Fatal("map map") }
    if FromAST(tr("set", false, false, tr("string", false, false))).String() != "set<string>" { t.Fatal("set map") }
    if FromAST(tr("slice", false, false, tr("int", false, false))).String() != "slice<int>" { t.Fatal("slice<T> map") }
    if FromAST(tr("T", true, false)).String() != "*T" { t.Fatal("ptr map") }
    if FromAST(tr("U", false, true)).String() != "[]U" { t.Fatal("slice [] map") }
}

