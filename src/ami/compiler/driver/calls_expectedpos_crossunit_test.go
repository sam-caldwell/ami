package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Ensure driver-provided paramPos is threaded into AnalyzeCallsWithSigs so that
// cross‑unit call diagnostics include expectedPos from the callee's declaration.
func TestCompile_CrossUnit_CallExpectedPos_FromCallee(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Place callee on line 3 to distinguish from call-site line numbers.
    fs.AddFile("u1.ami", "package app\n\nfunc Callee(a string, b int) {}\n")
    fs.AddFile("u2.ami", "package app\nfunc F(){ Callee(\"x\", \"y\") }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    // Find the mismatch for argIndex=1 and assert expectedPos is present and points to line 3
    found := false
    for _, d := range diags {
        if d.Code != "E_CALL_ARG_TYPE_MISMATCH" || d.Data == nil { continue }
        // ensure it's the second arg (index 1)
        var idx int
        if v, ok := d.Data["argIndex"].(int); ok { idx = v } else if vf, ok := d.Data["argIndex"].(float64); ok { idx = int(vf) }
        if idx != 1 { continue }
        ep, ok := d.Data["expectedPos"].(diag.Position)
        if !ok { t.Fatalf("expectedPos missing or wrong type: %+v", d) }
        if ep.Line != 3 {
            t.Fatalf("expectedPos line mismatch: got %d want 3 (diag=%+v)", ep.Line, d)
        }
        found = true
    }
    if !found { t.Fatalf("missing E_CALL_ARG_TYPE_MISMATCH with expectedPos for argIndex=1: %+v", diags) }
}

