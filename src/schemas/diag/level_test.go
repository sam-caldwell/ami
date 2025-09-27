package diag

import "testing"

func TestLevel_BasenamePair(t *testing.T) {
    if Info == "" || Warn == "" || Error == "" { t.Fatalf("empty constants") }
}

