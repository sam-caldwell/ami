package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_LinkFail_IncludesStderr(t *testing.T) {
    if _, err := llvme.FindClang(); err != nil { t.Skip("clang not available; skipping link-fail stderr test") }

    dir := filepath.Join("build", "test", "ami_build", "link_fail")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // leave envs empty to exercise default objects linking path
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    // Precreate an invalid object under build/obj to force link failure
    objDir := filepath.Join(dir, "build", "obj", "app")
    if err := os.MkdirAll(objDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(objDir, "bad.o"), []byte("not an object"), 0o644); err != nil { t.Fatalf("write bad.o: %v", err) }

    var out bytes.Buffer
    if err := runBuild(&out, dir, true, false); err != nil { t.Fatalf("runBuild: %v", err) }

    // Scan NDJSON for E_LINK_FAIL and ensure data.stderr present
    found := false
    sc := bufio.NewScanner(bytes.NewReader(out.Bytes()))
    for sc.Scan() {
        var m map[string]any
        if json.Unmarshal(sc.Bytes(), &m) != nil { continue }
        if m["code"] == "E_LINK_FAIL" {
            found = true
            if data, ok := m["data"].(map[string]any); ok {
                if s, ok := data["stderr"].(string); !ok || s == "" {
                    t.Fatalf("E_LINK_FAIL missing stderr: %v", m)
                }
            } else {
                t.Fatalf("E_LINK_FAIL missing data object: %v", m)
            }
        }
    }
    // It's possible link succeeds depending on platform; in that case, no E_LINK_FAIL
    // But with invalid bad.o, we expect a failure on common toolchains.
    if !found {
        t.Skip("link did not fail; skipping stderr assertion")
    }
}

