package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify directory layout is deterministic and mirrors package/unit structure; manifest paths are workspace‑relative.
func TestBuild_DirectoryLayout_DeterministicAndRelative(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "layout")
    _ = os.RemoveAll(dir)
    // workspace with simple unit
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // verbose build to emit debug artifacts
    if err := runBuild(os.Stdout, dir, false, true); err != nil { t.Fatalf("runBuild: %v", err) }
    // Directories by structure (use workspace package name)
    pkg := ws.Packages[0].Package.Name
    mustDir(t, filepath.Join(dir, "build", "debug", "ir", pkg))
    mustDir(t, filepath.Join(dir, "build", "debug", "asm", pkg))
    mustDir(t, filepath.Join(dir, "build", "obj", pkg))
    // Manifest paths are workspace‑relative
    b, err := os.ReadFile(filepath.Join(dir, "build", "ami.manifest"))
    if err != nil { t.Fatalf("read manifest: %v", err) }
    var m map[string]any
    if e := json.Unmarshal(b, &m); e != nil { t.Fatalf("json: %v; %s", e, string(b)) }
    for _, k := range []string{"objIndex", "debug"} {
        if v, ok := m[k].([]any); ok {
            for _, it := range v {
                s := it.(string)
                if strings.HasPrefix(s, "/") { t.Fatalf("%s not workspace‑relative: %s", k, s) }
                if !strings.HasPrefix(s, "build/") { t.Fatalf("%s not under build/: %s", k, s) }
            }
        }
    }
}

func mustDir(t *testing.T, p string) {
    t.Helper()
    st, err := os.Stat(p)
    if err != nil || !st.IsDir() { t.Fatalf("missing dir: %s (%v)", p, err) }
}
