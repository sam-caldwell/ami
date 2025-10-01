package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Cross-package variant: import + qualified call. Pending until package-qualified
// call resolution is implemented and SEM can look up signatures across packages.
func TestCompile_CrossPackage_CallExpectedPos_FromCallee(t *testing.T) {
    ws := workspace.Workspace{}

    // Library package with Callee defined on line 3
    libfs := &source.FileSet{}
    libfs.AddFile("lib1.ami", "package lib\n\nfunc Callee(a string, b int) {}\n")

    // App package importing lib and making a qualified call
    appfs := &source.FileSet{}
    appfs.AddFile("app1.ami", "package app\nimport lib\nfunc F(){ lib.Callee(\"x\", \"y\") }\n")

    pkgs := []Package{{Name: "lib", Files: libfs}, {Name: "app", Files: appfs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    // Expect E_CALL_ARG_TYPE_MISMATCH for argIndex=1 with expectedPos pointing to lib:3
    found := false
    for _, d := range diags {
        if d.Code != "E_CALL_ARG_TYPE_MISMATCH" || d.Data == nil { continue }
        var idx int
        if v, ok := d.Data["argIndex"].(int); ok { idx = v } else if vf, ok := d.Data["argIndex"].(float64); ok { idx = int(vf) }
        if idx != 1 { continue }
        if ep, ok := d.Data["expectedPos"].(diag.Position); ok {
            if ep.Line != 3 { t.Fatalf("expected lib Callee param line 3; got %d (diag=%+v)", ep.Line, d) }
            found = true
        }
    }
    if !found { t.Fatalf("missing cross-package expectedPos diag: %+v", diags) }
}
