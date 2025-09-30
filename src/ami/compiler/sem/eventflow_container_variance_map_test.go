package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEventFlow_ContainerVariance_MapKeyUnion_AllowsMember(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<map<string,int>>\"); B type(\"Event<map<Union<string,int>,int>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efcvmk1.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestEventFlow_ContainerVariance_MapKeyMismatch_Fails(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<map<string,string>>\"); B type(\"Event<map<int,string>>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efcvmk2.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    found := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { found = true; break } }
    if !found { t.Fatalf("expected mismatch for map key type") }
}

