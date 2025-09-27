package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_RemoteDepsOK_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "deps_ok")
    cache := filepath.Join(dir, "cache")
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir cache: %v", err) }
    // Point AMI_PACKAGE_CACHE to our test cache
    t.Setenv("AMI_PACKAGE_CACHE", cache)

    // Create a cached package with content and compute its hash
    pkgDir := filepath.Join(cache, "modA", "1.2.3")
    if err := os.MkdirAll(pkgDir, 0o755); err != nil { t.Fatalf("mkdir pkg: %v", err) }
    if err := os.WriteFile(filepath.Join(pkgDir, "x.txt"), []byte("hi"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    sha, err := workspace.HashDir(pkgDir)
    if err != nil { t.Fatalf("hash: %v", err) }

    // Workspace with remote requirement
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir dir: %v", err) }
    ws := workspace.DefaultWorkspace()
    p := ws.FindPackage("main")
    p.Import = []string{"modA 1.2.3"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save ws: %v", err) }

    // ami.sum matching the cache
    m := workspace.Manifest{Schema: "ami.sum/v1"}
    m.Set("modA", "1.2.3", sha)
    if err := m.Save(filepath.Join(dir, "ami.sum")); err != nil { t.Fatalf("save sum: %v", err) }

    var buf bytes.Buffer
    if err := runBuild(&buf, dir, true); err != nil { t.Fatalf("runBuild: %v", err) }
    var mres map[string]any
    if e := json.Unmarshal(buf.Bytes(), &mres); e != nil { t.Fatalf("json: %v; out=%s", e, buf.String()) }
    if mres["code"] != "BUILD_OK" { t.Fatalf("expected BUILD_OK; got %v", mres["code"]) }
}

