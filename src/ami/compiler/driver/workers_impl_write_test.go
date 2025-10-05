package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func Test_writeWorkersImplC_WritesWrappers(t *testing.T) {
    pkg := "app"
    names := []string{"W", "my-worker"}
    path, err := writeWorkersImplC(pkg, names)
    if err != nil { t.Fatalf("write: %v", err) }
    if _, err := os.Stat(path); err != nil { t.Fatalf("stat: %v", err) }
    b, _ := os.ReadFile(path)
    s := string(b)
    if !strings.Contains(s, "ami_worker_W") || !strings.Contains(s, "ami_worker_my_worker") {
        t.Fatalf("missing wrapper symbols in output: %s", path)
    }
    // ensure path is under build/debug/ir/<pkg>
    if !strings.Contains(filepath.Clean(path), filepath.Join("build", "debug", "ir", pkg)) {
        t.Fatalf("unexpected path: %s", path)
    }
}

