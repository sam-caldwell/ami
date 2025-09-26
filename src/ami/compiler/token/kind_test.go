package token

import "testing"

func TestKind_Enum_BasicOrdering(t *testing.T) {
    if EOF != 0 {
        t.Fatalf("EOF should be iota start (0), got %d", EOF)
    }
    if IDENT <= ILLEGAL {
        t.Fatalf("IDENT should come after ILLEGAL: %d %d", IDENT, ILLEGAL)
    }
    if KW_PACKAGE <= STRING {
        t.Fatalf("KW_PACKAGE should come after STRING: %d %d", KW_PACKAGE, STRING)
    }
}

