package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEventFlow_Union_ExpectedContainsUpstream_OK(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<string>\"); B type(\"Event<Union<int,string>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("ef1.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestEventFlow_Union_ExpectedNotContainingUpstream_Fails(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<string>\"); B type(\"Event<Union<int,int64>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("ef2.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    found := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { found = true; break } }
    if !found { t.Fatalf("expected mismatch for non-member union") }
}

func TestEventFlow_Optional_MatchInner_OK(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<Optional<int>>\"); B type(\"Event<Optional<int>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("ef3.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestEventFlow_Optional_DownstreamOptional_AllowsNonOptionalUpstream(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<int>\"); B type(\"Event<Optional<int>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("ef4.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestEventFlow_Optional_UpstreamOptional_DownstreamNonOptional_Fails(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<Optional<int>>\"); B type(\"Event<int>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("ef5.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    found := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { found = true; break } }
    if !found { t.Fatalf("expected mismatch for Optional upstream -> non-Optional downstream") }
}

func TestEventFlow_OptionalUnion_Downstream_AllowsMemberUpstream(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<string>\"); B type(\"Event<Optional<Union<int,string>>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("ef6.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestEventFlow_OptionalUnion_Downstream_MissingMember_Fails(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<string>\"); B type(\"Event<Optional<Union<int,int64>>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("ef7.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    found := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { found = true; break } }
    if !found { t.Fatalf("expected mismatch for Optional<Union> not containing upstream type") }
}
