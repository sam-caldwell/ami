package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_ManifestIncludesIntegrityEvidence(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "manifest_integrity")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(filepath.Join(dir, "cache", "lib", "1.0.0"), 0o755); err != nil { t.Fatalf("mkdir cache: %v", err) }
    // dummy cached content
    if err := os.WriteFile(filepath.Join(dir, "cache", "lib", "1.0.0", "x.txt"), []byte("ok"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    sha, err := workspace.HashDir(filepath.Join(dir, "cache", "lib", "1.0.0"))
    if err != nil { t.Fatalf("hash: %v", err) }
    // sum file matching cache
    sum := []byte("{\n  \"schema\": \"ami.sum/v1\",\n  \"packages\": {\n    \"lib\": { \"1.0.0\": \"" + sha + "\" }\n  }\n}\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }
    // workspace
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // point cache env
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", filepath.Join(dir, "cache"))
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    b, err := os.ReadFile(filepath.Join(dir, "build", "ami.manifest"))
    if err != nil { t.Fatalf("read: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    integ, ok := m["integrity"].(map[string]any)
    if !ok { t.Fatalf("integrity missing: %v", m) }
    ver, ok := integ["verified"].([]any)
    if !ok || len(ver) == 0 { t.Fatalf("verified missing: %v", integ) }
}

