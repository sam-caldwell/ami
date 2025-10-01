package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure struct field paths are included in E_GENERIC_ARITY_MISMATCH at call sites.
func TestCalls_StructFieldPath_ArityMismatch(t *testing.T) {
    // H expects Struct{a: slice<Owned<T>>}; local var provides Struct{a: slice<Owned<int,string>>}
    src := "package app\n" +
        "func H(p Struct{a:slice<Owned<T>>}){}\n" +
        "func F(){ var y Struct{a:slice<Owned<int,string>>}; H(y) }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeCalls(af)
    found := false
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" && d.Data != nil {
            if fp, ok := d.Data["fieldPath"].([]string); ok {
                if len(fp) >= 1 && fp[0] == "a" { found = true; break }
            }
        }
    }
    if !found { t.Fatalf("expected fieldPath with Struct->a; got %+v", ds) }
}
