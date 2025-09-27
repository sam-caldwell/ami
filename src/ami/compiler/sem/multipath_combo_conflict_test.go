package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestMergeAttr_PartitionKeyConflict(t *testing.T) {
    code := "package app\npipeline P() { Collect merge.PartitionBy(\"p\"), merge.Key(\"k\"); egress }\n"
    f := (&source.FileSet{}).AddFile("mpk.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeMultiPath(af)
    has := false
    for _, d := range ds { if d.Code == "E_MERGE_ATTR_CONFLICT" { has = true } }
    if !has { t.Fatalf("expected E_MERGE_ATTR_CONFLICT, got %v", ds) }
}

