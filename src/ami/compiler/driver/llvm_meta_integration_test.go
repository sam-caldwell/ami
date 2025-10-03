package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Verify the compiled .ll contains the AMI module metadata global when the program declares
// directives and a pipeline with collect/merge.
func TestCompile_EmbedsModuleMeta_LLVM(t *testing.T) {
    ws := workspace.Workspace{}
    var fs source.FileSet
    code := "package app\n" +
        "#pragma concurrency level=3\n" +
        "#pragma backpressure policy=block\n" +
        "#pragma schedule policy=fair\n" +
        "#pragma capabilities list=io,net\n" +
        "#pragma trust level=trusted\n" +
        "pipeline P(){ Collect merge.Buffer(10, dropNewest), merge.Sort(\"ts\", asc), merge.Stable(); egress }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: &fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read ll: %v", err) }
    s := string(b)
    if !strings.Contains(s, "@ami_meta_json = private constant") {
        t.Fatalf("expected meta global in ll: %s", s)
    }
}

