package scheduler

import "testing"

func TestParsePolicy(t *testing.T) {
    if p, ok := ParsePolicy("fair"); !ok || p != FAIR { t.Fatalf("parse fair: %v %v", p, ok) }
    if p, ok := ParsePolicy("FIFO"); !ok || p != FIFO { t.Fatalf("parse FIFO: %v %v", p, ok) }
    if p, ok := ParsePolicy("work-steal"); !ok || p != WORKSTEAL { t.Fatalf("parse worksteal: %v %v", p, ok) }
    if _, ok := ParsePolicy("unknown"); ok { t.Fatalf("expected false for unknown policy") }
}

