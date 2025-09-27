package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// In verbose mode, build writes debug artifacts and build/ami.manifest includes cross references under "debug".
func TestRunBuild_Verbose_BuildManifestDebugRefs(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "verbose_manifest_debug")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runBuild(os.Stdout, dir, false, true); err != nil { t.Fatalf("runBuild: %v", err) }
    // build manifest contains debug references
    b, err := os.ReadFile(filepath.Join(dir, "build", "ami.manifest"))
    if err != nil { t.Fatalf("read: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    dbg, ok := m["debug"].([]any)
    if !ok || len(dbg) == 0 { t.Fatalf("debug refs missing: %v", m) }
}

