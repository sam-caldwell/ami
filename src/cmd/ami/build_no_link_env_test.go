package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_NoLinkEnv_SkipsPerEnvLinking(t *testing.T) {
    if _, err := llvme.FindClang(); err != nil { t.Skip("clang not available; skipping no-link-env test") }
    env := hostEnv()
    if env == "" { t.Skip("unsupported host env for test") }
    old := buildNoLinkEnvs
    buildNoLinkEnvs = []string{env}
    defer func(){ buildNoLinkEnvs = old }()

    dir := filepath.Join("build", "test", "ami_build", "no_link_env")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Compiler.Env = []string{env}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    // Create a dummy per-env object so linking would normally occur
    envObjDir := filepath.Join(dir, "build", env, "obj", "app")
    if err := os.MkdirAll(envObjDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(envObjDir, "u.o"), []byte("obj"), 0o644); err != nil { t.Fatalf("write o: %v", err) }
    if err := os.WriteFile(filepath.Join(envObjDir, "index.json"), []byte("{}"), 0o644); err != nil { t.Fatalf("write idx: %v", err) }

    if err := runBuild(os.Stdout, dir, true, false); err != nil { t.Fatalf("runBuild: %v", err) }

    // Verify manifest does not include binaries for this env
    p := filepath.Join(dir, "build", "ami.manifest")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    if be, _ := m["binariesByEnv"].(map[string]any); be != nil && be[env] != nil {
        t.Fatalf("expected no binaries for env %s when disabled: %v", env, be)
    }
}

// hostEnv is defined in build_link_per_env_test.go
