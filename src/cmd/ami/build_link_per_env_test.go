package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "runtime"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func hostEnv() string {
    switch runtime.GOOS + "/" + runtime.GOARCH {
    case "darwin/arm64":
        return "darwin/arm64"
    case "darwin/amd64":
        return "darwin/amd64"
    case "linux/amd64":
        return "linux/amd64"
    case "linux/arm64":
        return "linux/arm64"
    default:
        return ""
    }
}

func TestRunBuild_PerEnvLinking_WritesBinaryAndManifest(t *testing.T) {
    if _, err := llvme.FindClang(); err != nil { t.Skip("clang not available; skipping per-env link test") }
    env := hostEnv()
    if env == "" { t.Skip("unsupported host env for test") }

    dir := filepath.Join("build", "test", "ami_build", "link_env")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Compiler.Env = []string{env}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }

    // Verify manifest has binariesByEnv for our env
    p := filepath.Join(dir, "build", "ami.manifest")
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    be, _ := m["binariesByEnv"].(map[string]any)
    if be == nil || be[env] == nil { t.Fatalf("binariesByEnv missing %s: %v", env, m) }
}

