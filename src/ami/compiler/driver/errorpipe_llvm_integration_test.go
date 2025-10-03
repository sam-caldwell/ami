package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure that a pipeline with an error block results in a .ll artifact embedding
// error pipeline metadata as a global when EmitLLVMOnly is enabled.
func TestCompile_EmbedsErrorPipelineMetadata_LLVM(t *testing.T) {
    _ = t.TempDir() // sandbox path not strictly needed; build artifacts go under ./build
    // minimal workspace
    ws := workspace.Workspace{}
    // package with a single unit containing an error block
    src := "package app\n" +
        "pipeline P(){ error { ingress.Transform().egress } }\n"
    var fs source.FileSet
    _ = fs.AddFile(filepath.Join(".", "u.ami"), src)
    pkgs := []Package{{Name: "app", Files: &fs}}
    // compile with EmitLLVMOnly to write .ll under build/debug/llvm/app/u.ll
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read ll: %v", err) }
    s := string(b)
    if !strings.Contains(s, "@ami_errpipe_P = private constant") {
        t.Fatalf("expected error pipeline global in ll: %s", s)
    }
}
