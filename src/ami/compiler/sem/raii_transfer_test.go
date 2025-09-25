package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

// Ownership transfer to a callee that accepts Owned<T> should satisfy RAII.
func TestRAII_OwnedParam_Transfer_OK(t *testing.T) {
    src := `package p
func consume(o Owned<string>) {}
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { mutate(consume(r)) }
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics { if d.Code == "E_RAII_OWNED_NOT_RELEASED" { t.Fatalf("unexpected not-released: %v", d) } }
}
