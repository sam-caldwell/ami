package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeDecorators_BuiltinsAndResolution(t *testing.T) {
    src := "package app\n@deprecated(\"msg\")\n@metrics\n@Helper\nfunc F(){}\nfunc Helper(){}\n"
    f := &source.File{Name: "d.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeDecorators(af)
    if len(ds) != 0 { t.Fatalf("unexpected diags: %+v", ds) }
}

func TestAnalyzeDecorators_Unresolved_And_Conflict(t *testing.T) {
    src := "package app\n@unknown\n@dec(1)\n@dec(2)\nfunc F(){}\n"
    f := &source.File{Name: "d2.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeDecorators(af)
    var unresolved, conflict bool
    for _, d := range ds {
        if d.Code == "E_DECORATOR_UNRESOLVED" { unresolved = true }
        if d.Code == "E_DECORATOR_CONFLICT" { conflict = true }
    }
    if !unresolved || !conflict { t.Fatalf("expected unresolved+conflict: %+v", ds) }
}

