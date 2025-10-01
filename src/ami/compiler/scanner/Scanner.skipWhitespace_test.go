package scanner

import "testing"

func TestScanner_skipWhitespace(t *testing.T) {
    s := &Scanner{file: nil, offset: 0}
    src := " \t\n x"
    n := len(src)
    s.skipWhitespace(src, n)
    if s.offset == 0 { t.Fatalf("expected offset advanced past whitespace") }
}

