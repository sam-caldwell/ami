package driver

import (
    "os"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestSourcesDebug_Write_EmptyFile(t *testing.T) {
    p, err := writeSourcesDebug("main", "unit", &ast.File{})
    if err != nil { t.Fatalf("write: %v", err) }
    if _, err := os.Stat(p); err != nil { t.Fatalf("stat: %v", err) }
}

