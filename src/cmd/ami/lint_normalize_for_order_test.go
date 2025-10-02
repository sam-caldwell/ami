package main

import "testing"

func TestNormalizeForOrder_FilePair(t *testing.T) {
    _ = normalizeForOrder([]string{"./A/", "b"})
}

