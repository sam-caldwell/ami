package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Multi-result functions are supported; ensure no E_LLVM_EMIT diagnostic is produced.
func TestCompile_DoesNotEmit_LLVMEmit_OnMultiResult(t *testing.T) {
    fs := &source.FileSet{}
    // Multi-result function triggers emitter scaffold error
    code := "package app\nfunc F()(int,int){ return 1,2 }\n"
    fs.AddFile("m.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(workspace.Workspace{}, pkgs, Options{Debug: false, EmitLLVMOnly: false})
    for _, d := range diags {
        if d.Code == "E_LLVM_EMIT" {
            t.Fatalf("did not expect E_LLVM_EMIT for multi-result function; got diags=%+v", diags)
        }
    }
}
