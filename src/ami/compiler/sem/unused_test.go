package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeUnused_VarsAndFuncs(t *testing.T) {
    code := "package app\nfunc used(){}\nfunc unused(){}\nfunc main(){ var x int; var y int; used(); _ = x }\n"
    f := (&source.FileSet{}).AddFile("u.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeUnused(af)
    var sawUnusedVarY, sawUnusedFunc bool
    for _, d := range ds {
        if d.Code == "W_UNUSED_VAR" && d.Message == "unused variable: y" { sawUnusedVarY = true }
        if d.Code == "W_UNUSED_FUNC" && d.Message == "unused function: unused" { sawUnusedFunc = true }
    }
    if !sawUnusedVarY || !sawUnusedFunc { t.Fatalf("expected both unused warnings; got: %+v", ds) }
}

