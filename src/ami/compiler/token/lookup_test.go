package token

import "testing"

func TestLookupKeyword_FindsAndFallsBack(t *testing.T) {
    if k, ok := LookupKeyword("pipeline"); !ok || k == Ident {
        t.Fatalf("expected pipeline keyword, got %v ok=%v", k, ok)
    }
    if k, ok := LookupKeyword("notakeyword"); ok || k != Ident {
        t.Fatalf("expected Ident fallback, got %v ok=%v", k, ok)
    }
}

