package ascii

import "testing"

func TestWrapLine_FilePair(t *testing.T) {
    got := wrapLine("abcdef", 3)
    if got != "abc\ndef" { t.Fatalf("wrapLine: %q", got) }
}

