package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestPackageImport_Validation_Good(t *testing.T) {
    f := (&source.FileSet{}).AddFile("ok.ami", "package app\nimport x.y.z\nfunc F(){}\n")
    p := parser.New(f)
    af, _ := p.ParseFile()
    diags := AnalyzePackageAndImports(af)
    if len(diags) != 0 { t.Fatalf("unexpected diags: %+v", diags) }
}

func TestPackageImport_Validation_Bad(t *testing.T) {
    f := (&source.FileSet{}).AddFile("bad.ami", "package bad_name\nimport \"weird space\"\nimport \"/abs\"\nimport \"x//y\"\nimport \"x/../z\"\nimport \"a/\"\n")
    p := parser.New(f)
    af, _ := p.ParseFile()
    diags := AnalyzePackageAndImports(af)
    if len(diags) < 5 { t.Fatalf("expected multiple diags, got %d: %+v", len(diags), diags) }
}
