package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_WritesBuildManifest(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "write_manifest")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    // build/ami.manifest exists and contains schema + objIndex
    p := filepath.Join(dir, "build", "ami.manifest")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    if m["schema"] != "ami.manifest/v1" { t.Fatalf("schema: %v", m["schema"]) }
    if _, ok := m["objIndex"].([]any); !ok { t.Fatalf("objIndex missing: %v", m) }
}

