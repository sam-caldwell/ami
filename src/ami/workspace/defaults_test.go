package workspace

import "testing"

func TestDefaultWorkspace_BasenamePair(t *testing.T) {
    w := DefaultWorkspace()
    if w.Version == "" || w.Packages == nil { t.Fatalf("unexpected default: %+v", w) }
}

