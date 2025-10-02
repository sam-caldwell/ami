package codegen

import "testing"

func TestSelectDefaultBackend_Unknown(t *testing.T) {
    if err := SelectDefaultBackend("__nosuch__"); err == nil { t.Fatalf("expected error for unknown backend") }
}

