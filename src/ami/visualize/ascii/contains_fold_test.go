package ascii

import "testing"

func TestContainsFold_FilePair(t *testing.T) {
    if !containsFold("Hello", "he") { t.Fatalf("expected true") }
}

