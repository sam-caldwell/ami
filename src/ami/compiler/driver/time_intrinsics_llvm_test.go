package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify Now/Add/Delta/Unix/UnixNano lower to corresponding runtime externs and calls.
func TestLower_Time_Intrinsics_Emit_All_Externs_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport time\nfunc F(){ time.Now(); var t = time.Now(); time.Add(t, 5); time.Delta(t, t); time.Unix(t); time.UnixNano(t) }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    decls := []string{
        "declare i64 @ami_rt_time_now()",
        "declare i64 @ami_rt_time_add(i64, i64)",
        "declare i64 @ami_rt_time_delta(i64, i64)",
        "declare i64 @ami_rt_time_unix(i64)",
        "declare i64 @ami_rt_time_unix_nano(i64)",
    }
    for _, d := range decls { if !strings.Contains(s, d) { t.Fatalf("missing extern: %s\n%s", d, s) } }
    calls := []string{
        "call i64 @ami_rt_time_now()",
        "call i64 @ami_rt_time_add(",
        "call i64 @ami_rt_time_delta(",
        "call i64 @ami_rt_time_unix(",
        "call i64 @ami_rt_time_unix_nano(",
    }
    for _, c := range calls { if !strings.Contains(s, c) { t.Fatalf("missing call: %s\n%s", c, s) } }
}
