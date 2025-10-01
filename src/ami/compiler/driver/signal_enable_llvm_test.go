package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure signal.Enable lowers to ami_rt_os_signal_enable with mapped signal number and emits extern.
func TestLower_Signal_Enable_Emits_OS_Enable_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nimport signal\nfunc F(){ signal.Enable(signal.SIGTERM) }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare void @ami_rt_os_signal_enable(i64)") {
        t.Fatalf("missing extern for os_signal_enable:\n%s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_os_signal_enable(i64 15)") {
        t.Fatalf("missing call for os_signal_enable SIGTERM=15:\n%s", s)
    }
}

