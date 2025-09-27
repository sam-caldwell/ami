package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_Verbose_PlanIncludesObjIndex(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "plan_objindex")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // Write a simple source to trigger compile and obj emission
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runBuild(os.Stdout, dir, false, true); err != nil { t.Fatalf("runBuild: %v", err) }
    p := filepath.Join(dir, "build", "debug", "build.plan.json")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if e := json.Unmarshal(b, &obj); e != nil { t.Fatalf("json: %v", e) }
    if _, ok := obj["objIndex"].([]any); !ok { t.Fatalf("objIndex missing or wrong type: %T", obj["objIndex"]) }
}

