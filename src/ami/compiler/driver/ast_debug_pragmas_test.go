package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestASTDebug_IncludesPragmas(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n#pragma concurrency level=4\nfunc F(){}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ast", "app", "u.ast.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ast: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    pr, ok := obj["pragmas"].([]any)
    if !ok || len(pr) != 1 { t.Fatalf("pragmas: %T len=%d", obj["pragmas"], len(pr)) }
    p0 := pr[0].(map[string]any)
    if p0["domain"] != "concurrency" { t.Fatalf("domain: %v", p0["domain"]) }
    if m, ok := p0["params"].(map[string]any); !ok || m["level"].(string) == "" {
        t.Fatalf("params: %#v", p0["params"])
    }
}
