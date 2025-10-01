package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Ensure return diagnostics include expectedPos pointing to the declared result type
// in cross-unit scenarios (callee in another unit within the same package).
func TestCompile_CrossUnit_ReturnExpectedPos_FromDecl(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Place F on line 3; expectedPos should reference this line for result index 1.
    fs.AddFile("u1.ami", "package app\n\nfunc F() (Owned<T>, Error<E>) { return Producer() }\n")
    fs.AddFile("u2.ami", "package app\nfunc Producer() (Owned<int,string>, Error<int,string>) { return }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    found := false
    for _, d := range diags {
        if d.Code != "E_GENERIC_ARITY_MISMATCH" || d.Data == nil { continue }
        // second result position
        var idx int
        if v, ok := d.Data["index"].(int); ok { idx = v } else if vf, ok := d.Data["index"].(float64); ok { idx = int(vf) }
        if idx != 1 { continue }
        ep, ok := d.Data["expectedPos"].(diag.Position)
        if !ok { t.Fatalf("expectedPos missing or wrong type: %+v", d) }
        if ep.Line != 3 { t.Fatalf("expectedPos line mismatch: got %d want 3 (diag=%+v)", ep.Line, d) }
        found = true
    }
    if !found { t.Fatalf("missing return E_GENERIC_ARITY_MISMATCH with expectedPos for index=1: %+v", diags) }
}

