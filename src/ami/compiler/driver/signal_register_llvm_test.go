package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure signal.Register lowers to ami_rt_signal_register and declares extern with (i64, i64).
func TestLower_Signal_Register_Emits_Extern_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport signal\nfunc H(){}\nfunc F(){ signal.Register(signal.SIGINT, H) }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare void @ami_rt_signal_register(i64, i64)") {
        t.Fatalf("missing signal_register extern in LLVM: %s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_signal_register(") {
        t.Fatalf("missing signal_register call in LLVM: %s", s)
    }
}

