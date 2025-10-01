package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestScanDurationLiteral_SimpleAndCompound(t *testing.T) {
    cases := []string{"300ms", "5s", "2h45m", "1.5h"}
    for _, c := range cases {
        s := New(&source.File{Name: "t", Content: c})
        if lex, ok := s.scanDurationLiteral(); !ok || lex != c {
            t.Fatalf("scan %q => ok=%v lex=%q", c, ok, lex)
        }
    }
}

func TestScanDurationLiteral_InvalidFraction(t *testing.T) {
    s := New(&source.File{Name: "t", Content: "1."})
    if _, ok := s.scanDurationLiteral(); ok { t.Fatalf("expected invalid for '1.'") }
}
