package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestMemoryDomains_CrossRefIntoState_Error(t *testing.T) {
    src := `package p
func f(ctx Context, ev Event<string>, r Owned<string>, st *State) Event<string> { mut { *st = &ev } }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_CROSS_DOMAIN_REF" || d.Code == "E_DEREF_UNSAFE" { found = true; break } }
    if !found { t.Fatalf("expected a cross-domain or unsafe-deref diagnostic; diags=%v", res.Diagnostics) }
}
