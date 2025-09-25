package mod

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadSum_InvalidJSON_ReturnsError(t *testing.T) {
    path := filepath.Join(t.TempDir(), "ami.sum")
    if err := os.WriteFile(path, []byte("{"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if _, err := LoadSumForCLI(path); err == nil {
        t.Fatalf("expected error for invalid JSON")
    }
}

