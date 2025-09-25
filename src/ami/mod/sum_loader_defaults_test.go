package mod

import (
	"path/filepath"
	"testing"
)

func TestLoadSum_DefaultsOnMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ami.sum")
	s, err := LoadSumForCLI(path)
	if err != nil {
		t.Fatalf("LoadSumForCLI: %v", err)
	}
	if s.Schema == "" || s.Packages == nil {
		t.Fatalf("expected defaults for empty sum: %+v", s)
	}
	if len(s.Packages) != 0 {
		t.Fatalf("expected empty packages")
	}
}
