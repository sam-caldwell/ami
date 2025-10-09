package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure Owned<slice<uint8>> participates in RAII lowering (zeroization) in debug LLVM.
func TestRAII_Bufio_OwnedSlice_Lowering_EmitsZeroize(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Simulate an Owned<slice<uint8>> value and explicit release(a)
    // This exercises the Owned RAII path independent of specific bufio calls,
    // but mirrors the expected return type from bufio.Reader.Read.
    code := "package app\nfunc F(){ var a Owned<slice<uint8>>; release(a) }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare void @ami_rt_zeroize_owned(ptr)") {
        t.Fatalf("missing zeroize_owned extern in LLVM: %s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_zeroize_owned(") {
        t.Fatalf("missing zeroize_owned call in LLVM: %s", s)
    }
}

