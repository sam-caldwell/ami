package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestAnalyzeMultiPath_FieldNameValidation(t *testing.T) {
    code := "package app\npipeline P(){ Collect merge.Sort(\"1bad\"), merge.Key(\"good_name\"), merge.PartitionBy(\"bad-name\"); egress }\n"
    f := (&source.FileSet{}).AddFile("mf.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeMultiPath(af)
    var invalid int
    for _, d := range ds { if d.Code == "E_MERGE_FIELD_NAME_INVALID" { invalid++ } }
    if invalid < 2 { t.Fatalf("expected at least two invalid field diagnostics; got %d (%+v)", invalid, ds) }
}

