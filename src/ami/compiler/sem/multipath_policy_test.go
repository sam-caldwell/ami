package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestMergeBuffer_InvalidPolicy_Error(t *testing.T) {
    code := "package app\npipeline P(){ Collect merge.Buffer(10, invalidPolicy); egress }\n"
    f := (&source.FileSet{}).AddFile("mpol.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeMultiPath(af)
    has := false
    for _, d := range ds { if d.Code == "E_MERGE_ATTR_ARGS" { has = true } }
    if !has { t.Fatalf("expected E_MERGE_ATTR_ARGS for invalid policy: %+v", ds) }
}

func TestMergeSort_DuplicateSame_NoConflict(t *testing.T) {
    code := "package app\npipeline P(){ Collect merge.Sort(\"ts\", asc), merge.Sort(\"ts\", asc); egress }\n"
    f := (&source.FileSet{}).AddFile("mdedup.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeMultiPath(af)
    for _, d := range ds { if d.Code == "E_MERGE_ATTR_CONFLICT" { t.Fatalf("unexpected conflict: %+v", ds) } }
}

