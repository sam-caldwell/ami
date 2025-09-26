package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Ensure legacy 3-param worker signature is accepted and does not raise E_WORKER_SIGNATURE.
func TestWorkers_LegacySignature_Accepted_WithDeprecatedInfo(t *testing.T) {
    src := `package p
func f(ctx Context, ev Event<int>, st State) Event<int> { }
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) }`
    p := parser.New(src)
    f := p.ParseFile()
    // Inspect declared function and assert legacy signature matches
    var fd astpkg.FuncDecl
    for _, d := range f.Decls {
        if fn, ok := d.(astpkg.FuncDecl); ok && fn.Name == "f" { fd = fn; break }
    }
    if fd.Name == "" { t.Fatalf("func decl f not found") }
    if !isLegacyWorkerSignature(fd) { t.Fatalf("expected isLegacyWorkerSignature to be true for f; fd=%+v", fd) }
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_WORKER_SIGNATURE" {
            t.Fatalf("unexpected signature error for legacy worker: %+v", d)
        }
    }
}
