package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_ManifestListsBinaries(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "manifest_bins")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // create a dummy executable before build
    if err := os.MkdirAll(filepath.Join(dir, "build"), 0o755); err != nil { t.Fatalf("mkdir build: %v", err) }
    bin := filepath.Join(dir, "build", "app.bin")
    if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil { t.Fatalf("write bin: %v", err) }
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    b, err := os.ReadFile(filepath.Join(dir, "build", "ami.manifest"))
    if err != nil { t.Fatalf("read: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    bins, ok := m["binaries"].([]any)
    if !ok || len(bins) == 0 { t.Fatalf("binaries missing or empty: %v", m) }
}

