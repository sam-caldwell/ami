package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeMultiPath_StableWithoutSort_Warns(t *testing.T) {
    src := "package app\npipeline P(){ Collect merge.Stable(); egress }\n"
    f := &source.File{Name: "mp_stable.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeMultiPath(af)
    has := false
    for _, d := range ds { if d.Code == "W_MERGE_STABLE_WITHOUT_SORT" { has = true } }
    if !has { t.Fatalf("expected W_MERGE_STABLE_WITHOUT_SORT: %+v", ds) }
}

