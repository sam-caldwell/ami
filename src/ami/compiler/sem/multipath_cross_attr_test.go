package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// When both merge.Key and merge.Sort are specified, recommend sorting by key.
func TestAnalyzeMultiPath_SortShouldIncludeKey_Warns(t *testing.T) {
    code := "package app\n" +
        "pipeline P(){ Collect merge.Key(\"id\"), merge.Sort(\"ts\"); egress }\n"
    f := (&source.FileSet{}).AddFile("mp_key_sort_warn.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeMultiPath(af)
    has := false
    for _, d := range ds { if d.Code == "W_MERGE_SORT_NOT_BY_KEY" { has = true } }
    if !has { t.Fatalf("expected W_MERGE_SORT_NOT_BY_KEY, got %+v", ds) }
}

