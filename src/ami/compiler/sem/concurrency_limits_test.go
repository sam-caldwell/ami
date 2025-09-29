package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestAnalyzeConcurrency_Limits_InvalidAndUnknown(t *testing.T) {
    code := "package app\n#pragma concurrency:limits ingress=0 foo=3 transform=2\n"
    f := (&source.FileSet{}).AddFile("cl.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeConcurrency(af)
    var bad, unk bool
    for _, d := range ds { if d.Code == "E_CONCURRENCY_LIMITS_INVALID" { bad = true }; if d.Code == "E_CONCURRENCY_LIMITS_KEY_UNKNOWN" { unk = true } }
    if !bad || !unk { t.Fatalf("expected both invalid and unknown diagnostics: %+v", ds) }
}

