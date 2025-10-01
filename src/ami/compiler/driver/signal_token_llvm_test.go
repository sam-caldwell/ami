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

// Ensure signal.Token(H) returns a literal i64 path in LLVM via add i64 0, <tok>.
func TestLower_Signal_Token_Emits_Literal_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport signal\nfunc H(){}\nfunc F() (int64) { return signal.Token(H) }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    h := fnv.New64a(); _, _ = h.Write([]byte("H"))
    tok := int64(h.Sum64())
    wantFrag := "add i64 0, " + strconv.FormatInt(tok, 10)
    if !strings.Contains(s, wantFrag) {
        t.Fatalf("missing token literal emission: %s\n%s", wantFrag, s)
    }
}
