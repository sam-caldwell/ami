package driver

import (
    "os"
    "testing"
)

func TestEventMetaDebug_Write_Empty(t *testing.T) {
    p, err := writeEventMetaDebug("pkg", "u")
    if err != nil { t.Fatalf("write: %v", err) }
    if _, err := os.Stat(p); err != nil { t.Fatalf("stat: %v", err) }
}

