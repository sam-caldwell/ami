package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "os"
    "path/filepath"
)

func TestHasUserMain_DetectsMainFunc(t *testing.T) {
    tmp := t.TempDir()
    // Create a fake workspace with one main package under ./app
    ws := workspace.Workspace{
        Packages: workspace.PackageList{
            {Key: "main", Package: workspace.Package{Name: "main", Root: "app"}},
        },
    }
    appDir := filepath.Join(tmp, "app")
    if err := os.MkdirAll(appDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Minimal AMI file with a main function
    src := "package main\nfunc main() {}\n"
    if err := os.WriteFile(filepath.Join(appDir, "main.ami"), []byte(src), 0o644); err != nil {
        t.Fatalf("write: %v", err)
    }
    if !hasUserMain(ws, tmp) { t.Fatalf("expected hasUserMain to return true") }
}
