package workspace

import "testing"

func TestManifest_Validate_Empty(t *testing.T) {
    var m Manifest
    v, miss, mis, err := m.Validate()
    if err != nil { t.Fatalf("Validate error: %v", err) }
    if len(v) != 0 || len(miss) != 0 || len(mis) != 0 {
        t.Fatalf("expected empty results, got v=%v miss=%v mis=%v", v, miss, mis)
    }
}

