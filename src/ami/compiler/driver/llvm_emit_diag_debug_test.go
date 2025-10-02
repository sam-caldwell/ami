package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// In debug mode, multi-result emission should succeed; ensure no E_LLVM_EMIT.
func TestCompile_Debug_No_LLVMEmit_OnMultiResult(t *testing.T) {
    fs := &source.FileSet{}
    code := "package app\nfunc F()(int,int){ return 1,2 }\n"
    fs.AddFile("m2.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(workspace.Workspace{}, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    for _, d := range diags {
        if d.Code == "E_LLVM_EMIT" {
            t.Fatalf("did not expect E_LLVM_EMIT in debug diagnostics for multi-result; got %+v", diags)
        }
    }
}
