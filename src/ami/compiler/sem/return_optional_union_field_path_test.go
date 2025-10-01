package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Optional inside struct: ensure fieldPath and path are stable in arity mismatch.
func TestReturn_Optional_StructFieldPath_ArityMismatch(t *testing.T) {
    code := "package app\n" +
        "func F() (Struct{a:Optional<Struct{b:slice<Owned<T>>}>}) { var x Struct{a:Optional<Struct{b:slice<Owned<int,string>>}>}; return x }\n"
    var fs source.FileSet
    f := fs.AddFile("opt_ret.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeReturnTypes(af)
    found := false
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" && d.Data != nil {
            if fp, ok := d.Data["fieldPath"].([]string); ok {
                if len(fp) >= 2 && fp[0] == "a" && fp[1] == "b" { found = true; break }
            }
        }
    }
    if !found { t.Fatalf("expected fieldPath a->b for Optional nested struct; got %+v", ds) }
}

// Union containing a struct arm with the mismatch; ensure fieldPath and path remain usable.
func TestReturn_Union_StructFieldPath_ArityMismatch(t *testing.T) {
    code := "package app\n" +
        "func F() (Struct{a:Union<Struct{b:slice<Owned<T>>},Struct{c:int}>}) { var x Struct{a:Union<Struct{b:slice<Owned<int,string>>},Struct{c:int}>}; return x }\n"
    var fs source.FileSet
    f := fs.AddFile("union_ret.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeReturnTypes(af)
    found := false
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" && d.Data != nil {
            // At minimum, expect path to include Owned and fieldPath to include a,b
            okFP := false
            if fp, ok := d.Data["fieldPath"].([]string); ok { if len(fp) >= 2 && fp[0] == "a" && fp[1] == "b" { okFP = true } }
            okPath := false
            if pth, ok := d.Data["path"].([]string); ok { if len(pth) >= 1 && pth[len(pth)-1] == "Owned" { okPath = true } }
            if okFP && okPath { found = true; break }
        }
    }
    if !found { t.Fatalf("expected fieldPath a->b and path ending with Owned for Union; got %+v", ds) }
}

