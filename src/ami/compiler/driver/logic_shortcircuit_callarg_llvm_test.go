package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLower_Logic_CallArg_ShortCircuitLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nfunc G(x bool){}\nfunc F(){ var a bool; var b bool; G(a || b) }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "br i1 ") { t.Fatalf("missing cond br for || arg:\n%s", s) }
    if !strings.Contains(s, "call void @G") { t.Fatalf("missing call to G:\n%s", s) }
}

