package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Selector receiver: ensure a chained receiver compiles to the Unix runtime call.
func TestLower_Time_Methods_Selector_Receiver_Unix_Emit_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // s.a is a selector receiver; we don't assign a real Time field but verify lowering emits the call.
    src := "package app\nimport time\nfunc F(){ var s Struct{a:Time}; s.a.Unix() }\n"
    fs.AddFile("sm.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "sm.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "declare i64 @ami_rt_time_unix(i64)") {
        t.Fatalf("missing extern for unix: \n%s", s)
    }
    if !strings.Contains(s, "call i64 @ami_rt_time_unix(") {
        t.Fatalf("missing call for unix: \n%s", s)
    }
}

