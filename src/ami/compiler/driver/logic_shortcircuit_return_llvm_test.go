package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLower_Logic_And_Or_Return_ShortCircuitLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nfunc F() (bool){ var a bool; var b bool; return a && b }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "br i1 ") { t.Fatalf("missing cond br for &&:\n%s", s) }
    if !strings.Contains(s, "phi i1") { t.Fatalf("missing phi i1 for && join:\n%s", s) }
}

