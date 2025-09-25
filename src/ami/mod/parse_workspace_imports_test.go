package mod

import (
    "os"
    "path/filepath"
    "testing"
)

func TestParseWorkspaceImports_ExtractsPaths(t *testing.T) {
    dir := t.TempDir()
    ws := filepath.Join(dir, "ami.workspace")
    content := `version: 1.0.0
packages:
  - main:
      version: 0.0.1
      root: ./src
      import:
        - ./a ==latest
        - ./b ^v1.2.3
`
    if err := os.WriteFile(ws, []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    imps, err := parseWorkspaceImports(ws)
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(imps) != 2 || imps[0] != "./a" || imps[1] != "./b" {
        t.Fatalf("unexpected imports: %v", imps)
    }
}

