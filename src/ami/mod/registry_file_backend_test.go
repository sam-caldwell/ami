package mod

import (
    "os"
    "path/filepath"
    "testing"
)

// TestFileBackend_Fetch_LocalPath copies a declared local path into the cache
// and returns no package/version information.
func TestFileBackend_Fetch_LocalPath(t *testing.T) {
    // Setup temporary workspace
    ws, err := os.MkdirTemp("", "ami-ws-")
    if err != nil { t.Fatalf("temp: %v", err) }
    defer os.RemoveAll(ws)
    // Create local project directory
    proj := filepath.Join(ws, "example")
    if err := os.MkdirAll(proj, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(proj, "README.md"), []byte("ok"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // ami.workspace with declared import
    wsFile := "packages:\n  import:\n    - ./example\n"
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsFile), 0o644); err != nil { t.Fatalf("workspace: %v", err) }
    // Change into workspace
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    if err := os.Chdir(ws); err != nil { t.Fatalf("chdir: %v", err) }

    dest, pkg, ver, err := GetWithInfo("./example")
    if err != nil { t.Fatalf("GetWithInfo: %v", err) }
    if pkg != "" || ver != "" { t.Fatalf("file backend should not return pkg/ver; got %q %q", pkg, ver) }
    // Check destination exists and contains README
    if _, err := os.Stat(filepath.Join(dest, "README.md")); err != nil {
        t.Fatalf("dest missing README: %v", err)
    }
}

