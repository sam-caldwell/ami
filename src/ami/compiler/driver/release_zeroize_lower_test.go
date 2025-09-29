package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Lowering release(x) emits a call to ami_rt_zeroize in the generated LLVM for debug builds.
func TestLower_Release_EmitsZeroizeCall_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Owned is treated as a handle; release(a) should emit zeroize call.
    code := "package app\nfunc F(){ var a Owned; release(a) }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare void @ami_rt_zeroize(ptr, i64)") {
        t.Fatalf("missing zeroize extern in LLVM: %s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_zeroize(") {
        t.Fatalf("missing zeroize call in LLVM: %s", s)
    }
}

