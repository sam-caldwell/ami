package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDeclaredLocalImport_NormalizesLeadingDotSlash(t *testing.T) {
	ws := t.TempDir()
	content := `version: 1.0.0
packages:
  - main:
      version: 0.0.1
      root: ./src
      import:
        - ./lib ==latest
`
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Query without leading ./ should still match
	ok, err := isDeclaredLocalImport(filepath.Join(ws, "ami.workspace"), "lib")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("expected declaration match without leading ./")
	}
}
