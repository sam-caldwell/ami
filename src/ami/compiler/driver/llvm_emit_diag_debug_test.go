package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure E_LLVM_EMIT is surfaced in debug mode as well when emitter fails.
func TestCompile_Debug_Emits_LLVMEmit_Diagnostic_OnEmitterFailure(t *testing.T) {
    fs := &source.FileSet{}
    code := "package app\nfunc F()(int,int){ return 1,2 }\n"
    fs.AddFile("m2.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(workspace.Workspace{}, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    found := false
    for _, d := range diags { if d.Code == "E_LLVM_EMIT" { found = true; break } }
    if !found { t.Fatalf("expected E_LLVM_EMIT in diagnostics (debug)") }
}

