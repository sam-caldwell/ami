package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestAnalyzeConcurrency_WorkersAndSchedule_Invalid(t *testing.T) {
    code := "package app\n#pragma concurrency:workers 0\n#pragma concurrency:schedule invalid\n"
    f := (&source.FileSet{}).AddFile("c.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeConcurrency(af)
    var badW, badS bool
    for _, d := range ds { if d.Code == "E_CONCURRENCY_WORKERS_INVALID" { badW = true }; if d.Code == "E_CONCURRENCY_SCHEDULE_INVALID" { badS = true } }
    if !badW || !badS { t.Fatalf("expected both invalid diagnostics: %+v", ds) }
}

