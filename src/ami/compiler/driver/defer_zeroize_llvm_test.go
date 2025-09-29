package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify that DEFER is emitted as a zeroize-owned call before function return in LLVM.
func TestLower_DeferRelease_EmitsZeroizeOwned_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ var a Owned; defer release(a) }\n"
    fs.AddFile("d2.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "d2.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare void @ami_rt_zeroize_owned(ptr)") {
        t.Fatalf("missing zeroize_owned extern: %s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_zeroize_owned(") {
        t.Fatalf("missing zeroize_owned call: %s", s)
    }
}

