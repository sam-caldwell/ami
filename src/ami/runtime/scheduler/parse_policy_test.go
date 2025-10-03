package scheduler

import "testing"

func TestParsePolicy_NormalizesKnownValues(t *testing.T) {
    cases := map[string]Policy{"fifo": FIFO, "FIFO": FIFO, "lifo": LIFO, "LIFO": LIFO, "fair": FAIR, "FAIR": FAIR, "worksteal": WORKSTEAL, "work-steal": WORKSTEAL}
    for in, want := range cases {
        if got, ok := ParsePolicy(in); !ok || got != want { t.Fatalf("in=%q got=%q ok=%v", in, got, ok) }
    }
    if _, ok := ParsePolicy("unknown"); ok { t.Fatalf("expected false for unknown") }
}
