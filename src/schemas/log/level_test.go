package logschema

import "testing"

func TestLevel_Constants(t *testing.T) {
    // compile-time constant presence and basic equality
    if Trace == "" || Debug == "" || Info == "" || Warn == "" || Error == "" || Fatal == "" {
        t.Fatalf("expected non-empty level constants")
    }
    if Info != Level("info") { t.Fatalf("unexpected value for Info: %q", Info) }
}

