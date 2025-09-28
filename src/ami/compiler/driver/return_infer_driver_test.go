package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure E_TYPE_UNINFERRED appears for unannotated functions with ambiguous returns.
func TestCompile_ReturnInference_Unannotated_Ambiguous(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // F has no declared results and returns an untyped local ident.
    // The local is declared without type or initializer, so it remains 'any'.
    // Return inference should report E_TYPE_UNINFERRED among diagnostics.
    fs.AddFile("u.ami", "package app\nfunc F(){ var a; return a }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    var hasUninferred bool
    for _, d := range diags { if d.Code == "E_TYPE_UNINFERRED" { hasUninferred = true; break } }
    if !hasUninferred {
        t.Fatalf("expected E_TYPE_UNINFERRED for ambiguous unannotated return; got: %+v", diags)
    }
}
