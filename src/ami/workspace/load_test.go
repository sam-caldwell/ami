package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func TestWorkspace_Load_Basic(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace", "load")
    _ = os.MkdirAll(dir, 0o755)
    path := filepath.Join(dir, "ami.workspace")
    w := DefaultWorkspace()
    if err := w.Save(path); err != nil { t.Fatalf("save: %v", err) }
    var got Workspace
    if err := got.Load(path); err != nil { t.Fatalf("load: %v", err) }
    if got.Version == "" || got.Toolchain.Compiler.Target == "" { t.Fatalf("unexpected: %+v", got) }
}

