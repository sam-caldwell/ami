package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure struct field paths are included in E_GENERIC_ARITY_MISMATCH at return sites.
func TestReturn_StructFieldPath_ArityMismatch(t *testing.T) {
    code := "package app\n" +
        "func F() (Struct{a:slice<Owned<T>>}) { var x Struct{a:slice<Owned<int,string>>}; return x }\n"
    var fs source.FileSet
    f := fs.AddFile("u.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeReturnTypes(af)
    found := false
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" && d.Data != nil {
            if fp, ok := d.Data["fieldPath"].([]string); ok {
                if len(fp) >= 2 && fp[0] == "Struct" && fp[1] == "a" { found = true; break }
            }
        }
    }
    if !found { t.Fatalf("expected fieldPath Struct->a in return diag; got %+v", ds) }
}
