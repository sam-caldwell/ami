package driver

import (
    "fmt"
    "hash/fnv"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    a "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// When the handler is not an ident or selector (e.g., a call expression),
// we fall back to a stable token based on the expression position (anon@<offset>).
func TestLower_Signal_Register_FallbackToken_ForNonIdent_InLLVM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Note: pass H() as the handler expression to trigger fallback tokenization.
    src := "package app\nimport signal\nfunc H(){}\nfunc F(){ signal.Register(signal.SIGINT, H()) }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    // Compute expected fallback token based on AST expr offset of handler arg (CallExpr).
    pr := parser.New(fs.Files[0])
    af, _ := pr.ParseFileCollect()
    if af == nil { t.Fatalf("parse returned nil AST") }
    // Find function F and get its signal.Register call; then take arg1 position.
    var off int = -1
    for _, d := range af.Decls {
        if fn, ok := d.(*a.FuncDecl); ok && fn.Name == "F" {
            if fn.Body == nil { t.Fatalf("missing body") }
            for _, st := range fn.Body.Stmts {
                if es, ok := st.(*a.ExprStmt); ok {
                    if ce, ok := es.X.(*a.CallExpr); ok {
                        // expect name like signal.Register
                        if strings.HasSuffix(ce.Name, ".Register") {
                            if len(ce.Args) >= 2 {
                                if hcall, ok := ce.Args[1].(*a.CallExpr); ok {
                                    off = hcall.Pos.Offset
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    if off < 0 { t.Fatalf("did not find handler call expr offset") }
    key := fmt.Sprintf("anon@%d", off)
    h := fnv.New64a(); _, _ = h.Write([]byte(key))
    tok := int64(h.Sum64())
    want := "call void @ami_rt_signal_register(i64 2, i64 " + strconv.FormatInt(tok, 10) + ")"
    if !strings.Contains(s, want) {
        t.Fatalf("missing fallback token immediate in call:\nwant contains: %s\n%s", want, s)
    }
}
