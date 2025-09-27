package workspace

import "testing"

func TestPackage_BasenamePair(t *testing.T) {
    p := Package{Name: "app", Version: "0.0.1", Root: "./src"}
    if p.Name == "" || p.Version == "" || p.Root == "" { t.Fatalf("unexpected: %+v", p) }
}

