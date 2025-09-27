package ir

import "testing"

func TestModule_BasenamePair(t *testing.T) {
    m := Module{Package: "main"}
    if m.Package != "main" { t.Fatalf("unexpected: %+v", m) }
}

