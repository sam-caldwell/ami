package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Cross-package variant: app returns lib.Producer(); return diagnostics should include
// expectedPos pointing to the app function's declared result type position.
func TestCompile_CrossPackage_ReturnExpectedPos_FromDecl(t *testing.T) {
    ws := workspace.Workspace{}

    // Library package with Producer
    libfs := &source.FileSet{}
    libfs.AddFile("lib1.ami", "package lib\nfunc Producer() (Owned<int,string>, Error<int,string>) { return }\n")

    // App package importing lib and returning lib.Producer()
    appfs := &source.FileSet{}
    // Place F on line 3; expectedPos should reference this line for result index 1.
    appfs.AddFile("app1.ami", "package app\nimport lib\nfunc F() (Owned<T>, Error<E>) { return lib.Producer() }\n")

    pkgs := []Package{{Name: "lib", Files: libfs}, {Name: "app", Files: appfs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    // Expect E_GENERIC_ARITY_MISMATCH for index=1 with expectedPos pointing to app:3
    found := false
    for _, d := range diags {
        if d.Code != "E_GENERIC_ARITY_MISMATCH" || d.Data == nil { continue }
        var idx int
        if v, ok := d.Data["index"].(int); ok { idx = v } else if vf, ok := d.Data["index"].(float64); ok { idx = int(vf) }
        if idx != 1 { continue }
        if ep, ok := d.Data["expectedPos"].(diag.Position); ok {
            if ep.Line != 3 { t.Fatalf("expected app F result line 3; got %d (diag=%+v)", ep.Line, d) }
            found = true
        }
    }
    if !found { t.Fatalf("missing cross-package return expectedPos diag: %+v", diags) }
}

