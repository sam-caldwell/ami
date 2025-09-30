package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Downstream expects Struct{a:int}; upstream provides Struct{a:Optional<int>} → accept by unwrapping in field context.
func TestEventFlow_Struct_FieldOptional_Upstream_OK(t *testing.T) {
    code := "package app\n"+
        "pipeline P(){ A type(\"Event<Struct{a:Optional<int>,b:string}>\"); B type(\"Event<Struct{a:int}>\"); A -> B; egress }\n"
    f := (&source.FileSet{}).AddFile("efsfopt1.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeEventTypeFlow(af)
    for _, d := range ds { if d.Code == "E_EVENT_TYPE_FLOW" { t.Fatalf("unexpected mismatch: %+v", ds) } }
}

