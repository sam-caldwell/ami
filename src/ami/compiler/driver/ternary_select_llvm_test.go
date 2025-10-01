package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure ternary conditional lowers to a LLVM select instruction.
func TestLower_Ternary_Emits_Select_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nfunc F() (int){ var a int; var b int; return (a == 1) ? b : 2 }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "select i1") {
        t.Fatalf("missing select instruction in LLVM output:\n%s", s)
    }
}

