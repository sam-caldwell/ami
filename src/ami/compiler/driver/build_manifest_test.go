package driver

import (
    "os"
    "testing"
)

func TestBuildManifest_Write_EmptyOK(t *testing.T) {
    // Even with empty packages, writer should emit a file with default schema
    p, err := writeBuildManifest(BuildManifest{})
    if err != nil { t.Fatalf("write: %v", err) }
    if _, err := os.Stat(p); err != nil { t.Fatalf("stat: %v", err) }
}

