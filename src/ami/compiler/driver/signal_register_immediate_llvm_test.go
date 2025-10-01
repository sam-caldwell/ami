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

// Assert handler token is emitted as an immediate i64 in the LLVM call.
func TestLower_Signal_Register_HandlerImmediate_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport signal\nfunc H(){}\nfunc F(){ signal.Register(signal.SIGTERM, H) }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    // compute expected token for handler name H
    h := fnv.New64a(); _, _ = h.Write([]byte("H"))
    tok := int64(h.Sum64())
    wantFrag := "call void @ami_rt_signal_register(i64 15, i64 " + strconv.FormatInt(tok, 10) + ")"
    if !strings.Contains(s, wantFrag) {
        t.Fatalf("missing immediate handler token in call:\nwant contains: %s\n%s", wantFrag, s)
    }
}

