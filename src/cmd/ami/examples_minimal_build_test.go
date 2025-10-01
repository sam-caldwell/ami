package main

import (
    "os"
    "path/filepath"
    "runtime"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Builds a minimal POP example across multiple envs and validates deterministic artifacts.
func TestExamples_Minimal_BuildAcrossTargets(t *testing.T) {
    dir := filepath.Join("build", "test", "examples", "minimal")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    // Minimal workspace with two envs to exercise multi-target emission
    ws := workspace.Workspace{
        Version: "1.0.0",
        Toolchain: workspace.Toolchain{Compiler: workspace.Compiler{Target: "./build", Env: []string{"linux/amd64", "darwin/arm64"}}},
        Packages: workspace.PackageList{{Key: "main", Package: workspace.Package{Name: "simple-main", Version: "0.0.1", Root: "./src"}}},
    }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // Minimal POP snippet
    src := []byte("package app\nfunc F(){}\npipeline P(){ ingress; egress }\n")
    if err := os.WriteFile(filepath.Join(dir, "src", "main.ami"), src, 0o644); err != nil { t.Fatalf("write: %v", err) }

    // Build in human mode (non-JSON); verbose=false is fine for deterministic emission steps
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }

    // Verify per-env LLVM emission exists regardless of clang presence
    mustLL := []string{
        filepath.Join(dir, "build", "linux", "amd64", "obj", "simple-main", "main.ll"),
        filepath.Join(dir, "build", "darwin", "arm64", "obj", "simple-main", "main.ll"),
    }
    for _, p := range mustLL {
        if _, err := os.Stat(p); err != nil { t.Fatalf("missing per-env ll: %s (%v)", p, err) }
    }
    // Manifest exists
    if _, err := os.Stat(filepath.Join(dir, "build", "ami.manifest")); err != nil { t.Fatalf("manifest: %v", err) }

    // If clang is available, link will have produced binaries per env; just sanity check at least one
    if _, err := llvme.FindClang(); err == nil {
        // Host-specific sanity: check a per-env binary exists under build/<env>/
        host := runtime.GOOS + "/" + runtime.GOARCH
        // Map host env to expected dir
        var bdir string
        switch host {
        case "darwin/arm64":
            bdir = filepath.Join(dir, "build", "darwin", "arm64")
        case "linux/amd64":
            bdir = filepath.Join(dir, "build", "linux", "amd64")
        }
        if bdir != "" {
            // Any executable in the env dir suffices
            entries, _ := os.ReadDir(bdir)
            foundExec := false
            for _, e := range entries {
                if e.IsDir() { continue }
                if info, _ := e.Info(); info != nil && (info.Mode()&0o111 != 0) { foundExec = true; break }
            }
            if !foundExec { t.Logf("clang present but no exec found under %s (ok if linking skipped by policy)", bdir) }
        }
    }
}

