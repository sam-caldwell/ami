package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_BitwiseAndShiftOperators(t *testing.T) {
	// Ensure '&', '^', '<<', '>>' are recognized via the Operators map.
	src := "& ^ << >>"
	s := New(&source.File{Name: "ops", Content: src})
	var seq []token.Kind
	for {
		tok := s.Next()
		seq = append(seq, tok.Kind)
		if tok.Kind == token.EOF {
			break
		}
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
