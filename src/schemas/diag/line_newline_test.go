package diag

import (
    "strings"
    "testing"
    "time"
)

func TestDiagLine_AppendsNewline(t *testing.T) {
    r := Record{Timestamp: time.Unix(0, 0), Level: Warn, Code: "W001", Message: "warn"}
    line := Line(r)
    if !strings.HasSuffix(string(line), "\n") {
        t.Fatalf("expected newline suffix")
    }
}

