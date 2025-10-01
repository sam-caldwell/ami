package driver

import (
    "hash/fnv"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify additional enum-to-OS mappings for SIGHUP and SIGQUIT.
func TestLower_Signal_Register_SIGHUP_SIGQUIT_Mapping_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nimport signal\nfunc H(){}\nfunc A(){ signal.Register(signal.SIGHUP, H) }\nfunc B(){ signal.Register(signal.SIGQUIT, H) }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    // token for H
    h := fnv.New64a(); _, _ = h.Write([]byte("H"))
    tok := int64(h.Sum64())
    wants := []string{
        "call void @ami_rt_signal_register(i64 1, i64 " + strconv.FormatInt(tok, 10) + ")",
        "call void @ami_rt_signal_register(i64 3, i64 " + strconv.FormatInt(tok, 10) + ")",
    }
    for _, w := range wants {
        if !strings.Contains(s, w) {
            t.Fatalf("missing mapping in call: %s\n%s", w, s)
        }
    }
}

