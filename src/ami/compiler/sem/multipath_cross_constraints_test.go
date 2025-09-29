package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeMultiPath_DedupWithoutField_RequiresKey_Warns(t *testing.T) {
    src := "package app\npipeline P(){ Collect merge.Dedup(); egress }\n"
    f := &source.File{Name: "mp_dedup.ami", Content: src}
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeMultiPath(af)
    has := false
    for _, d := range ds { if d.Code == "W_MERGE_DEDUP_WITHOUT_KEY" { has = true } }
    if !has { t.Fatalf("expected W_MERGE_DEDUP_WITHOUT_KEY, got %+v", ds) }
}

