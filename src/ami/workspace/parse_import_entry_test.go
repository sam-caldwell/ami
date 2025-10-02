package workspace

import "testing"

func TestParseImportEntry_Simple(t *testing.T) {
    p, c := ParseImportEntry("mod >= 1.2.3")
    if p != "mod" || c == "" { t.Fatalf("unexpected: %q %q", p, c) }
}

