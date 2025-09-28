package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Force an LLVM emitter failure (multi-result function) and ensure E_LLVM_EMIT appears in diagnostics.
func TestCompile_Emits_LLVMEmit_Diagnostic_OnEmitterFailure(t *testing.T) {
    fs := &source.FileSet{}
    // Multi-result function triggers emitter scaffold error
    code := "package app\nfunc F()(int,int){ return 1,2 }\n"
    fs.AddFile("m.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(workspace.Workspace{}, pkgs, Options{Debug: false, EmitLLVMOnly: false})
    if len(diags) == 0 { t.Fatalf("expected diagnostics") }
    found := false
    for _, d := range diags { if d.Code == "E_LLVM_EMIT" { found = true; break } }
    if !found { t.Fatalf("expected E_LLVM_EMIT in diagnostics: %+v", diags) }
}

