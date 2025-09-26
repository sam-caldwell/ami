package token

import "testing"

func TestSymbols_CoreRunes(t *testing.T) {
    if LexLf != '\n' || LexCr != '\r' || LexTab != '\t' {
        t.Fatalf("line rune constants invalid: LF=%q CR=%q TAB=%q", LexLf, LexCr, LexTab)
    }
    if LexBoolEQ != "==" || LexBoolNE != "!=" {
        t.Fatalf("bool operator strings invalid: %q %q", LexBoolEQ, LexBoolNE)
    }
}

