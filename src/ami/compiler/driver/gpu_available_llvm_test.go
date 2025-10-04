package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify gpu.MetalAvailable lowers to mask accessor and emits externs in LLVM.
func TestLower_GPU_MetalAvailable_UsesMaskAccessor(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport gpu\nfunc F(){ if gpu.MetalAvailable() { } }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare i1 @ami_rt_gpu_has(i64)") {
        t.Fatalf("missing ami_rt_gpu_has extern in LLVM:\n%s", s)
    }
    if !strings.Contains(s, "call i1 @ami_rt_gpu_has(i64 0)") {
        t.Fatalf("missing call to ami_rt_gpu_has bit0 (metal):\n%s", s)
    }
}

