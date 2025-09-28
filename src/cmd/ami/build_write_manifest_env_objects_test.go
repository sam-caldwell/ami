package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_Manifest_Includes_PerEnvObjectsAndIndex(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "manifest_env")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Compiler.Env = []string{"linux/amd64"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    // Precreate per-env object and index
    envObjDir := filepath.Join(dir, "build", "linux/amd64", "obj", "app")
    if err := os.MkdirAll(envObjDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(envObjDir, "u.o"), []byte("obj"), 0o644); err != nil { t.Fatalf("write o: %v", err) }
    if err := os.WriteFile(filepath.Join(envObjDir, "index.json"), []byte("{}"), 0o644); err != nil { t.Fatalf("write idx: %v", err) }

    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }

    // Read manifest
    p := filepath.Join(dir, "build", "ami.manifest")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    obe, _ := m["objectsByEnv"].(map[string]any)
    if obe == nil || obe["linux/amd64"] == nil { t.Fatalf("objectsByEnv missing linux/amd64: %v", m) }
    oie, _ := m["objIndexByEnv"].(map[string]any)
    if oie == nil || oie["linux/amd64"] == nil { t.Fatalf("objIndexByEnv missing linux/amd64: %v", m) }
}

