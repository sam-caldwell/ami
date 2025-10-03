package main

import "testing"

func TestAtoiSafe_Cases(t *testing.T) {
    if atoiSafe("123") != 123 { t.Fatalf("simple parse failed") }
    if atoiSafe("12x") != 12 { t.Fatalf("stop at non-digit failed") }
    if atoiSafe("") != 0 { t.Fatalf("empty string should be 0") }
    // clamp on very large numbers (stop loop)
    if atoiSafe("9999999999999") <= 0 { t.Fatalf("expected positive after clamp") }
}
