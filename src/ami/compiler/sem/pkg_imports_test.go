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
    f := (&source.FileSet{}).AddFile("bad.ami", "package bad_name\nimport \"weird space\"\n")
    p := parser.New(f)
    af, _ := p.ParseFile()
    diags := AnalyzePackageAndImports(af)
    if len(diags) < 2 { t.Fatalf("expected diags, got %d: %+v", len(diags), diags) }
}

