package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAST_Decorators_Appear_OnFuncDecl(t *testing.T) {
    src := "package app\n@metrics\nfunc F(){}\n"
    f := &source.File{Name: "a.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    var found bool
    for _, d := range af.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok && fn.Name == "F" {
            if len(fn.Decorators) == 0 { t.Fatalf("decorators missing on func") }
            found = true
        }
    }
    if !found { t.Fatalf("func F not found") }
}

// worker-like signature parsing will be validated in later phases;
// this phase verifies decorators attach to functions without types.
