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

// Verify selector-form handler (alias.H) emits immediate token in LLVM call.
func TestLower_Signal_Register_SelectorHandler_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    mod := &source.FileSet{}
    mod.AddFile("m.ami", "package m\nfunc H(){}\n")
    app := &source.FileSet{}
    app.AddFile("app.ami", "package app\nimport signal\nimport m\nfunc F(){ signal.Register(signal.SIGINT, m.H) }\n")
    pkgs := []Package{{Name: "m", Files: mod}, {Name: "app", Files: app}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "app.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    h := fnv.New64a(); _, _ = h.Write([]byte("m.H"))
    tok := int64(h.Sum64())
    wantFrag := "call void @ami_rt_signal_register(i64 2, i64 " + strconv.FormatInt(tok, 10) + ")"
    if !strings.Contains(s, wantFrag) {
        t.Fatalf("missing selector handler immediate token in call:\nwant contains: %s\n%s", wantFrag, s)
    }
}

