package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify that manifest includes artifacts[] entries with kind:"obj" when .o files exist.
func TestRunBuild_ManifestIncludesArtifacts(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "manifest_artifacts")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // Precreate a default object in build/obj/<pkg>
    objDir := filepath.Join(dir, "build", "obj", ws.Packages[0].Package.Name)
    if err := os.MkdirAll(objDir, 0o755); err != nil { t.Fatalf("mkdir obj: %v", err) }
    if err := os.WriteFile(filepath.Join(objDir, "unit.o"), []byte{0}, 0o644); err != nil { t.Fatalf("write .o: %v", err) }

    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    // Read manifest
    b, err := os.ReadFile(filepath.Join(dir, "build", "ami.manifest"))
    if err != nil { t.Fatalf("read manifest: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    arts, _ := m["artifacts"].([]any)
    if arts == nil || len(arts) == 0 { t.Fatalf("artifacts missing: %v", m) }
    // Verify at least one entry has kind:"obj"
    ok := false
    for _, a := range arts {
        if mm, ok2 := a.(map[string]any); ok2 {
            if mm["kind"] == "obj" { ok = true; break }
        }
    }
    if !ok { t.Fatalf("no artifact kind=obj present: %v", m) }
}

