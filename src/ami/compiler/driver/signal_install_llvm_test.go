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

// Ensure signal.Install(H) lowers to ami_rt_install_handler_thunk with immediate token and @H pointer.
func TestLower_Signal_Install_Emits_ThunkInstall_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport signal\nfunc H(){}\nfunc F(){ signal.Install(H) }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    // token for H
    h := fnv.New64a(); _, _ = h.Write([]byte("H"))
    tok := int64(h.Sum64())
    wantDecl := "declare void @ami_rt_install_handler_thunk(i64, ptr)"
    if !strings.Contains(s, wantDecl) { t.Fatalf("missing extern: %s\n%s", wantDecl, s) }
    wantCall := "call void @ami_rt_install_handler_thunk(i64 " + strconv.FormatInt(tok, 10) + ", ptr @H)"
    if !strings.Contains(s, wantCall) { t.Fatalf("missing install call: %s\n%s", wantCall, s) }
}

