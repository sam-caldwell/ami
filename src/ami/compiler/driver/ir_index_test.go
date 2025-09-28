package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestWriteIRIndex_PackageLevel(t *testing.T) {
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ var x int; x = x }\nfunc G(){ }\n"
    fs.AddFile("i.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(workspace.Workspace{}, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    // Expect ir.index.json under build/debug/ir/app
    p := filepath.Join("build", "debug", "ir", "app", "ir.index.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read ir.index.json: %v", err) }
    var obj map[string]any
    if e := json.Unmarshal(b, &obj); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    if obj["schema"] != "ir.index.v1" { t.Fatalf("schema: %v", obj["schema"]) }
    if obj["package"] != "app" { t.Fatalf("package: %v", obj["package"]) }
    if _, ok := obj["units"].([]any); !ok { t.Fatalf("units missing: %v", obj) }
}

