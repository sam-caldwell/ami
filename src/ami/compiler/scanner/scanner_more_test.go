package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_PunctuationBracketsAndSymbols(t *testing.T) {
    // Cover LBracket/RBracket, Backslash, Dollar, Tick, Tilde, Question, At, SingleQuote, and fallback Symbol.
    src := "[ ] \\ $ ` ~ ? @ ' §"
    s := New(&source.File{Name: "t", Content: src})
    kinds := []token.Kind{}
    for {
        tok := s.Next()
        kinds = append(kinds, tok.Kind)
        if tok.Kind == token.EOF { break }
    }
    // Note: fallback Symbol on a multi-byte rune like '§' will produce one Symbol per byte.
    want := []token.Kind{
        token.LBracketSym,
        token.RBracketSym,
        token.BackslashSym,
        token.DollarSym,
        token.TickSym,
        token.TildeSym,
        token.QuestionSym,
        token.AtSym,
        token.SingleQuoteSym,
        token.Symbol, // '§' byte 1
        token.Symbol, // '§' byte 2 (fallback increments by 1 byte)
        token.EOF,
    }
    if len(kinds) != len(want) {
        t.Fatalf("token count mismatch: got %d want %d (%v)", len(kinds), len(want), kinds)
    }
    for i := range want {
        if kinds[i] != want[i] {
            t.Fatalf("kind[%d]=%v want %v (all=%v)", i, kinds[i], want[i], kinds)
        }
    }
}

func TestScanner_BitwiseAndShiftOperators(t *testing.T) {
    // Ensure '&', '^', '<<', '>>' are recognized via the Operators map.
    src := "& ^ << >>"
    s := New(&source.File{Name: "ops", Content: src})
    seq := []token.Kind{}
    for {
        tok := s.Next()
        seq = append(seq, tok.Kind)
        if tok.Kind == token.EOF { break }
    }
    want := []token.Kind{token.BitAnd, token.BitXor, token.Shl, token.Shr, token.EOF}
    if len(seq) != len(want) {
        t.Fatalf("len=%d want=%d (%v)", len(seq), len(want), seq)
    }
    for i := range want {
        if seq[i] != want[i] {
            t.Fatalf("seq[%d]=%v want %v (all=%v)", i, seq[i], want[i], seq)
        }
    }
}
