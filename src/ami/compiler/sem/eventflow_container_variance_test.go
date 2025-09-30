package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEventFlow_ContainerVariance_SliceOptionalUnion_AllowsMember(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<slice<string>>\"); B type(\"Event<slice<Optional<Union<int,string>>>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efcv1.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestEventFlow_ContainerVariance_MapValueUnion_AllowsMember(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<map<string,string>>\"); B type(\"Event<map<string,Union<int,string>>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efcv2.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestEventFlow_ContainerVariance_SliceMismatch_Fails(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<slice<string>>\"); B type(\"Event<slice<int>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efcv3.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    found := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { found = true; break } }
    if !found { t.Fatalf("expected mismatch for slice element type") }
}

func TestEventFlow_ContainerVariance_SetOptional_AllowsInner(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<set<int>>\"); B type(\"Event<set<Optional<int>>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efcv4.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

