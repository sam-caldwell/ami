package logging

import "testing"

func TestNormalizeMsg_CRLFtoLF(t *testing.T) {
    if got := normalizeMsg("a\r\nb"); got != "a\nb" { t.Fatalf("normalize: %q", got) }
}

