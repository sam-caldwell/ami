package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEventFlow_Struct_Width_Subtyping_OK(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<Struct{a:int,b:string}>\"); B type(\"Event<Struct{a:int}>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efsw1.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

func TestEventFlow_Struct_Width_Subtyping_Mismatch(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<Struct{a:int,b:int}>\"); B type(\"Event<Struct{a:string}>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efsw2.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    found := false
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { found = true; break } }
    if !found { t.Fatalf("expected mismatch for incompatible field type") }
}

