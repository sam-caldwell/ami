package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Using AMI stdlib time stubs via builtin bundle, ensure time.Sleep lowers to ami_rt_sleep_ms and declares extern.
func TestLower_Time_Sleep_Emits_SleepExtern_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport time\nfunc F(){ time.Sleep(10ms) }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare void @ami_rt_sleep_ms(i64)") {
        t.Fatalf("missing sleep_ms extern in LLVM: %s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_sleep_ms(") {
        t.Fatalf("missing sleep_ms call in LLVM: %s", s)
    }
}
