package main

import (
    "os"
    "path/filepath"
    "testing"
)

func TestRuntime_KV_Pragma_EmitsArtifacts(t *testing.T) {
    dir := t.TempDir()
    content := `#pragma test:case c1
#pragma test:runtime input={}
#pragma test:kv ns="p1/n1" put="a=1;b=2" get="a,b" emit=true
`
    if err := os.WriteFile(filepath.Join(dir, "kv_test.ami"), []byte(content), 0o644); err != nil { t.Fatal(err) }
    if err := runTest(os.Stdout, dir, false, false, 0); err != nil { t.Fatalf("runTest: %v", err) }
    // Verify artifacts exist
    base := filepath.Join(dir, "build", "test", "kv", "kv_test.ami_c1")
    if _, err := os.Stat(base + ".metrics.json"); err != nil { t.Fatalf("metrics missing: %v", err) }
    if _, err := os.Stat(base + ".dump.json"); err != nil { t.Fatalf("dump missing: %v", err) }
}

