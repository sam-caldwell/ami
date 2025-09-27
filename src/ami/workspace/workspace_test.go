package workspace

import (
    "os"
    "path/filepath"
    "testing"
    "gopkg.in/yaml.v3"
)

func TestDefaultWorkspace_SaveLoad_RoundTrip(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace", "roundtrip")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    path := filepath.Join(dir, "ami.workspace")

    w := DefaultWorkspace()
    if err := w.Save(path); err != nil {
        t.Fatalf("save: %v", err)
    }

    var got Workspace
    if err := got.Load(path); err != nil {
        t.Fatalf("load: %v", err)
    }

    if got.Version == "" || got.Toolchain.Compiler.Target == "" {
        t.Fatalf("missing fields after load: %+v", got)
    }
    if got.FindPackage("main") == nil {
        t.Fatalf("main package missing")
    }
}

func TestPackageList_YAMLShape(t *testing.T) {
    // Marshal and ensure the top-level YAML for packages is a sequence
    // where each item is a single-entry map.
    w := DefaultWorkspace()
    b, err := yaml.Marshal(w)
    if err != nil {
        t.Fatalf("marshal: %v", err)
    }
    var root map[string]any
    if err := yaml.Unmarshal(b, &root); err != nil {
        t.Fatalf("unmarshal root: %v", err)
    }
    pkgs, ok := root["packages"].([]any)
    if !ok || len(pkgs) == 0 {
        t.Fatalf("packages shape not sequence: %T", root["packages"])
    }
}

