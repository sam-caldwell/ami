package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestASTDebug_Positions_Exist(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nimport \"x\"\nfunc F(){}\npipeline P(){ Alpha() }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ast", "app", "u.ast.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ast: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    // funcs[0].pos present
    if fns, ok := obj["funcs"].([]any); ok && len(fns) > 0 {
        f0 := fns[0].(map[string]any)
        if _, ok := f0["pos"].(map[string]any); !ok { t.Fatalf("func pos missing: %#v", f0) }
    } else { t.Fatalf("funcs missing") }
    // pipelines[0].steps[0].pos present
    if pipes, ok := obj["pipelines"].([]any); ok && len(pipes) > 0 {
        p0 := pipes[0].(map[string]any)
        steps, _ := p0["steps"].([]any)
        if len(steps) == 0 { t.Fatalf("steps missing") }
        s0 := steps[0].(map[string]any)
        if _, ok := s0["pos"].(map[string]any); !ok { t.Fatalf("step pos missing: %#v", s0) }
    } else { t.Fatalf("pipelines missing") }
}

