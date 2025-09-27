package driver

import (
    "os"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestAsmDebug_Write_MinimalModule(t *testing.T) {
    m := ir.Module{}
    p, err := writeAsmDebug("main", "unit", nil, m)
    if err != nil { t.Fatalf("write: %v", err) }
    if _, err := os.Stat(p); err != nil { t.Fatalf("stat: %v", err) }
}

