package diag

import "testing"

func TestLine_AppendsNewline_BasenamePair(t *testing.T) {
    b := Line(Record{})
    if len(b) == 0 || b[len(b)-1] != '\n' { t.Fatalf("missing newline") }
}

