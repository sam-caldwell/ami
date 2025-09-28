package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure manifest includes env objects and indices for darwin/arm64 when present.
func TestRunBuild_WriteManifest_DarwinArm64_EnvObjects(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "manifest_env_darwin_arm64")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Compiler.Env = []string{"darwin/arm64"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    // Precreate per-env object and index
    envObjDir := filepath.Join(dir, "build", "darwin", "arm64", "obj", ws.Packages[0].Package.Name)
    if err := os.MkdirAll(envObjDir, 0o755); err != nil { t.Fatalf("mk env obj: %v", err) }
    if err := os.WriteFile(filepath.Join(envObjDir, "unit.o"), []byte{1}, 0o644); err != nil { t.Fatalf("write .o: %v", err) }
    if err := os.WriteFile(filepath.Join(envObjDir, "index.json"), []byte(`{"schema":"objindex.v1","package":"`+ws.Packages[0].Package.Name+`","units":[]}`), 0o644); err != nil { t.Fatalf("write index: %v", err) }

    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    mf := filepath.Join(dir, "build", "ami.manifest")
    b, err := os.ReadFile(mf)
    if err != nil { t.Fatalf("read manifest: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; body=%s", e, string(b)) }
    obe, _ := m["objectsByEnv"].(map[string]any)
    if obe == nil || obe["darwin/arm64"] == nil { t.Fatalf("objectsByEnv missing darwin/arm64: %v", m) }
    oie, _ := m["objIndexByEnv"].(map[string]any)
    if oie == nil || oie["darwin/arm64"] == nil { t.Fatalf("objIndexByEnv missing darwin/arm64: %v", m) }
}

