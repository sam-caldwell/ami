package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Unreferenced functions should be dropped from IR functions list (file-local DCE).
func TestCompile_IR_DCE_Removes_Unreferenced_Functions(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc used(){}\nfunc dead(){}\nfunc main(){ used() }\n"
    fs.AddFile("dce.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    arts, _ := Compile(ws, pkgs, Options{Debug: true})
    if len(arts.IR) == 0 { t.Fatalf("no IR emitted") }
    b, err := os.ReadFile(arts.IR[0])
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if e := json.Unmarshal(b, &obj); e != nil { t.Fatalf("json: %v", e) }
    fns := obj["functions"].([]any)
    // Ensure 'dead' not present
    for _, it := range fns {
        m := it.(map[string]any)
        if m["name"] == "dead" { t.Fatalf("dead should be DCE'd") }
    }
}
