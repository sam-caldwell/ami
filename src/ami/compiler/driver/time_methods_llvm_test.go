package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify t.Unix() and t.UnixNano() method calls lower to runtime externs and calls.
func TestLower_Time_Methods_Unix_UnixNano_Emit_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport time\nfunc F(){ var t = time.Now(); t.Unix(); t.UnixNano() }\n"
    fs.AddFile("m.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "m.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    // Ensure externs are present
    decls := []string{
        "declare i64 @ami_rt_time_unix(i64)",
        "declare i64 @ami_rt_time_unix_nano(i64)",
    }
    for _, d := range decls { if !strings.Contains(s, d) { t.Fatalf("missing extern: %s\n%s", d, s) } }
    // Ensure calls are present
    calls := []string{
        "call i64 @ami_rt_time_unix(",
        "call i64 @ami_rt_time_unix_nano(",
    }
    for _, c := range calls { if !strings.Contains(s, c) { t.Fatalf("missing call: %s\n%s", c, s) } }
}

