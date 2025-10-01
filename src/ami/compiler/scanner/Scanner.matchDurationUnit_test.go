package scanner

import "testing"

func TestMatchDurationUnit(t *testing.T) {
    // 2-char units
    if n := matchDurationUnit("ns", 0); n != 2 { t.Fatalf("ns => %d", n) }
    if n := matchDurationUnit("us", 0); n != 2 { t.Fatalf("us => %d", n) }
    if n := matchDurationUnit("ms", 0); n != 2 { t.Fatalf("ms => %d", n) }
    // 1-char units
    if n := matchDurationUnit("h", 0); n != 1 { t.Fatalf("h => %d", n) }
    if n := matchDurationUnit("m", 0); n != 1 { t.Fatalf("m => %d", n) }
    if n := matchDurationUnit("s", 0); n != 1 { t.Fatalf("s => %d", n) }
    // unknown
    if n := matchDurationUnit("xz", 0); n != 0 { t.Fatalf("xz => %d", n) }
}
