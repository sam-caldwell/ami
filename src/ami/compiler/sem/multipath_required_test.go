package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeMultiPath_RequiredFields(t *testing.T) {
    // merge.Sort without field → W_MERGE_SORT_NO_FIELD; empty field → E_MERGE_ATTR_REQUIRED
    src := "package app\npipeline P(){ Collect merge.Sort(); Collect merge.Sort(\"\"); egress }\n"
    f := &source.File{Name: "mp_req.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeMultiPath(af)
    var hasWarn, hasErr bool
    for _, d := range ds {
        if d.Code == "W_MERGE_SORT_NO_FIELD" { hasWarn = true }
        if d.Code == "E_MERGE_ATTR_REQUIRED" { hasErr = true }
    }
    if !hasWarn || !hasErr { t.Fatalf("expected warn+err; got %+v", ds) }
}

func TestAnalyzeMultiPath_WindowTinyAndDropAlias(t *testing.T) {
    src := "package app\npipeline P(){ Collect merge.Buffer(1, dropNewest), merge.Buffer(2, drop); egress }\n"
    f := &source.File{Name: "mp_buf.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeMultiPath(af)
    var tiny, alias bool
    for _, d := range ds {
        if d.Code == "W_MERGE_TINY_BUFFER" { tiny = true }
        if d.Code == "W_MERGE_BUFFER_DROP_ALIAS" { alias = true }
    }
    if !tiny || !alias { t.Fatalf("expected tiny+alias warnings; got %+v", ds) }
}

