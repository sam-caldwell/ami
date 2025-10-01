package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure fieldPath captures nested struct-of-struct field names at call sites.
func TestCalls_DeepStructFieldPath_ArityMismatch(t *testing.T) {
    // H expects Struct{a: Struct{b: slice<Owned<T>>}}; local var provides Struct{a: Struct{b: slice<Owned<int,string>>}}
    src := "package app\n" +
        "func H(p Struct{a:Struct{b:slice<Owned<T>>}}){}\n" +
        "func F(){ var y Struct{a:Struct{b:slice<Owned<int,string>>}}; H(y) }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeCalls(af)
    found := false
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" && d.Data != nil {
            if fp, ok := d.Data["fieldPath"].([]string); ok {
                if len(fp) >= 4 && fp[0] == "Struct" && fp[1] == "a" && fp[2] == "Struct" && fp[3] == "b" {
                    found = true
                    break
                }
            }
        }
    }
    if !found { t.Fatalf("expected fieldPath Struct->a->Struct->b; got %+v", ds) }
}

