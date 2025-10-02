package workspace

import "testing"

func TestCrossCheckRequirementsIntegrity_Empty(t *testing.T) {
    var m Manifest
    miss, mis, err := CrossCheckRequirementsIntegrity(&m, nil)
    if err != nil || len(miss) != 0 || len(mis) != 0 { t.Fatalf("unexpected: %v %v %v", miss, mis, err) }
}

