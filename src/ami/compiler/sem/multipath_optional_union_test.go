package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestMergeField_Optional_Primitive_Orderable(t *testing.T) {
    code := "package app\n" +
        "pipeline P(){ A type(\"Event<Struct{x:Optional<int>}>\"); Collect merge.Sort(\"x\"); A -> Collect; egress }\n"
    f := (&source.FileSet{}).AddFile("mopt1.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeMergeFieldTypes(af)
    if len(ds) != 0 { t.Fatalf("unexpected diagnostics: %+v", ds) }
}

func TestMergeField_Optional_Struct_Unorderable(t *testing.T) {
    code := "package app\n" +
        "pipeline P(){ A type(\"Event<Struct{x:Optional<Struct{k:int}>}>\"); Collect merge.Sort(\"x\"); A -> Collect; egress }\n"
    f := (&source.FileSet{}).AddFile("mopt2.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeMergeFieldTypes(af)
    found := false
    for _, d := range ds { if d.Code == "E_MERGE_SORT_FIELD_UNORDERABLE" { found = true; break } }
    if !found { t.Fatalf("expected unorderable for Optional<Struct>") }
}

func TestMergeField_Union_Primitives_Orderable(t *testing.T) {
    code := "package app\n" +
        "pipeline P(){ A type(\"Event<Struct{k:Union<int,string>}>\"); Collect merge.Sort(\"k\"); A -> Collect; egress }\n"
    f := (&source.FileSet{}).AddFile("mun1.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeMergeFieldTypes(af)
    if len(ds) != 0 { t.Fatalf("unexpected diagnostics: %+v", ds) }
}

func TestMergeField_Union_Mixed_Unorderable(t *testing.T) {
    code := "package app\n" +
        "pipeline P(){ A type(\"Event<Struct{k:Union<int,Struct{a:int}>}>\"); Collect merge.Sort(\"k\"); A -> Collect; egress }\n"
    f := (&source.FileSet{}).AddFile("mun2.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeMergeFieldTypes(af)
    found := false
    for _, d := range ds { if d.Code == "E_MERGE_SORT_FIELD_UNORDERABLE" { found = true; break } }
    if !found { t.Fatalf("expected unorderable for mixed union") }
}

func TestMergeField_Union_MissingField_Unknown(t *testing.T) {
    // Union of structs where only one has the field should count as unknown
    code := "package app\n" +
        "pipeline P(){ A type(\"Event<Union<Struct{k:int},Struct{m:int}>>\"); Collect merge.Sort(\"k\"); A -> Collect; egress }\n"
    f := (&source.FileSet{}).AddFile("mun3.ami", code)
    af, _ := parser.New(f).ParseFile()
    ds := AnalyzeMergeFieldTypes(af)
    found := false
    for _, d := range ds { if d.Code == "E_MERGE_SORT_FIELD_UNKNOWN" { found = true; break } }
    if !found { t.Fatalf("expected unknown field for union missing member") }
}
