package sem

import (
    "testing"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

// Ensure we emit a specific semantic error for pointer State parameters to guide users
// toward ambient state.get/set usage, even though the parser also flags pointer syntax.
func TestSem_StatePointerParam_Rejected(t *testing.T) {
    src := "package p\nfunc f(ctx Context, ev Event<string>, st *State) Event<string> { ev }\n"
    p := parser.New(src)
    f := p.ParseFile()
    if f == nil {
        t.Fatalf("nil AST")
    }
    // Expect either parser or semantic layer to reject *State param explicitly
    found := false
    for _, d := range p.Errors() { if d.Code == "E_STATE_PARAM_POINTER" { found = true; break } }
    if !found {
        diags := Check(f)
        for _, d := range diags { if d.Code == "E_STATE_PARAM_POINTER" { found = true; break } }
    }
    if !found { t.Fatalf("expected E_STATE_PARAM_POINTER from parser/sem; none found") }
}

// Sanity: no error when not using pointer state param
func TestSem_StateParam_NonPointer_Ok(t *testing.T) {
    src := "package p\nfunc f(ctx Context, ev Event<string>, st State) Event<string> { ev }\n"
    p := parser.New(src)
    f := p.ParseFile()
    if f == nil {
        t.Fatalf("nil AST")
    }
    diags := Check(f)
    for _, d := range diags {
        if d.Code == "E_STATE_PARAM_POINTER" {
            t.Fatalf("unexpected E_STATE_PARAM_POINTER for non-pointer state param: %v", diags)
        }
    }
    // Ensure the AST contains at least one func decl for coverage
    if len(f.Decls) == 0 {
        t.Fatalf("no decls parsed (unexpected)")
    }
    _ = astpkg.File{Decls: f.Decls}
}
