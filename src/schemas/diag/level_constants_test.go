package diag

import "testing"

func TestLevel_Constants(t *testing.T) {
    if Info != Level("info") || Warn != Level("warn") || Error != Level("error") {
        t.Fatalf("level constants mismatch: %q %q %q", Info, Warn, Error)
    }
}

