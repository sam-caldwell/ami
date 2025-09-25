package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Validate_Defaults(t *testing.T) {
	dir := t.TempDir()
	wsPath := filepath.Join(dir, "ami.workspace")
	content := "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env: []\n  linker: {}\n  linter: {}\npackages: []\n"
	if err := os.WriteFile(wsPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	ws, err := Load(wsPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(ws.Toolchain.Compiler.Env) == 0 {
		t.Fatalf("expected default env")
	}
}
