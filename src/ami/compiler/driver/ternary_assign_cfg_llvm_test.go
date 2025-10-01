package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure conditional assignment lowers to control flow with a conditional branch.
func TestLower_Ternary_Assign_Emits_CondBr_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nfunc F(){ var a int; var b int; a = (a == 1) ? b : 2 }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "br i1 ") {
        t.Fatalf("expected conditional branch in LLVM output:\n%s", s)
    }
    if !strings.Contains(s, "; assign %a = ") {
        t.Fatalf("expected assignment in blocks:\n%s", s)
    }
}

