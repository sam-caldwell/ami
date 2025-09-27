package ast

import "testing"

func TestFuncDecl_Basics(t *testing.T) {
    fn := &FuncDecl{Name: "main"}
    if fn.Name != "main" { t.Fatalf("name: %q", fn.Name) }
}

