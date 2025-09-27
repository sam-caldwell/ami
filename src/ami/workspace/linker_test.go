package workspace

import "testing"

func TestLinker_BasenamePair(t *testing.T) {
    l := Linker{}
    if l.Options != nil { t.Fatalf("unexpected: %+v", l) }
}

