package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestModUpdate_Human_PrintsSelected(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_update", "human_selected")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "a.txt"), []byte("a"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages = workspace.PackageList{{Key: "main", Package: workspace.Package{Name: "app", Version: "1.0.0", Root: "./src", Import: []string{"lib@^1.0.0"}}}}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // Provide ami.sum with versions for lib
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), []byte(`{"schema":"ami.sum/v1","packages":{"lib":{"v1.2.3":"a","v1.4.0":"b"}}}`), 0o644); err != nil { t.Fatalf("write sum: %v", err) }
    var buf bytes.Buffer
    if err := runModUpdate(&buf, dir, false); err != nil { t.Fatalf("runModUpdate: %v", err) }
    if !bytes.Contains(buf.Bytes(), []byte("select lib@v1.4.0")) {
        t.Fatalf("expected human selection output; got: %s", buf.String())
    }
}

