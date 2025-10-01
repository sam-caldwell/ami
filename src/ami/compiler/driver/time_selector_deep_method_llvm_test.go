package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Deep selector receiver: x.a.b.UnixNano()
func TestLower_Time_Methods_Selector_Deep_Receiver_UnixNano_Emit_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport time\nfunc F(){ var x Struct{a:Struct{b:Time}}; x.a.b.UnixNano() }\n"
    fs.AddFile("smd.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "smd.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare i64 @ami_rt_time_unix_nano(i64)") {
        t.Fatalf("missing extern for unix_nano: \n%s", s)
    }
    if !strings.Contains(s, "call i64 @ami_rt_time_unix_nano(") {
        t.Fatalf("missing call for unix_nano: \n%s", s)
    }
}

