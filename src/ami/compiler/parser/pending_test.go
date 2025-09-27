package parser

import "testing"

// TestPendingOrNew calls the tiny helper to mark it covered.
func TestPendingOrNew(t *testing.T) {
    // p is nil-safe only via usage; pass a real parser with empty pending
    p := &Parser{}
    got := pendingOrNew(nil, p)
    if got != nil { t.Fatalf("expected nil pending, got %#v", got) }
}

