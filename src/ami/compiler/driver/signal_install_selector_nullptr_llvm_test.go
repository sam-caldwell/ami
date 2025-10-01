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

// Cross-module handler pointers: ensure selector-form Install does not try to take a cross-module pointer.
// We still emit the deterministic token, but use a null pointer for the function symbol.
func TestLower_Signal_Install_Selector_NullPtr_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    mod := &source.FileSet{}
    mod.AddFile("m.ami", "package m\nfunc H(){}\n")
    app := &source.FileSet{}
    app.AddFile("app.ami", "package app\nimport signal\nimport m\nfunc F(){ signal.Install(m.H) }\n")
    pkgs := []Package{{Name: "m", Files: mod}, {Name: "app", Files: app}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "app.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    // token for m.H
    h := fnv.New64a(); _, _ = h.Write([]byte("m.H"))
    tok := int64(h.Sum64())
    want := "call void @ami_rt_install_handler_thunk(i64 " + strconv.FormatInt(tok, 10) + ", ptr null)"
    if !strings.Contains(s, want) {
        t.Fatalf("expected selector Install to use null ptr:\nwant contains: %s\n%s", want, s)
    }
}

