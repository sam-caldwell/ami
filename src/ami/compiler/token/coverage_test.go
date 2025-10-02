package token

import "testing"

// TestKind_String_AllCases ensures every defined Kind has a stable String()
// representation. This exercises all switch branches to improve coverage.
func testKind_String_AllCases(t *testing.T) {
	cases := map[Kind]string{
		Unknown:      "Unknown",
		EOF:          "EOF",
		Ident:        "Ident",
		Number:       "Number",
		String:       "String",
		Symbol:       "Symbol",
		Assign:       "Assign",
		Plus:         "Plus",
		Minus:        "Minus",
		Star:         "Star",
		Slash:        "Slash",
		Percent:      "Percent",
		Bang:         "Bang",
		Eq:           "Eq",
		Ne:           "Ne",
		Lt:           "Lt",
		Gt:           "Gt",
		Le:           "Le",
		Ge:           "Ge",
		And:          "And",
		Or:           "Or",
		Arrow:        "Arrow",
		KwPackage:    "KwPackage",
		KwImport:     "KwImport",
		KwFunc:       "KwFunc",
		KwReturn:     "KwReturn",
		KwPipeline:   "KwPipeline",
		KwIngress:    "KwIngress",
		KwEgress:     "KwEgress",
		KwError:      "KwError",
		LParenSym:    "LParen",
		RParenSym:    "RParen",
		LBraceSym:    "LBrace",
		RBraceSym:    "RBrace",
		LBracketSym:  "LBracket",
		RBracketSym:  "RBracket",
		CommaSym:     "Comma",
		SemiSym:      "Semi",
		DotSym:       "Dot",
		ColonSym:     "Colon",
		LineComment:  "LineComment",
		BlockComment: "BlockComment",
	}
	for k, want := range cases {
		if got := k.String(); got != want {
			t.Fatalf("Kind(%d).String() => %q; want %q", k, got, want)
		}
	}
	// Unknown fallback for out-of-range kinds
	if got := Kind(-1).String(); got != "Unknown" {
		t.Fatalf("Kind(-1).String() => %q; want Unknown", got)
	}
}

// TestPrecedence_AllOperators exercises every defined operator precedence branch.
func testPrecedence_AllOperators(t *testing.T) {
	// Or/And
	if Precedence(Or) == 0 || !(Precedence(Or) < Precedence(And)) {
		t.Fatalf("expected Or < And and both > 0")
	}
	// Eq/Ne vs Relational
	if !(Precedence(Eq) == Precedence(Ne)) {
		t.Fatalf("expected Eq and Ne to have same precedence")
	}
	if !(Precedence(Eq) < Precedence(Lt)) || !(Precedence(Le) == Precedence(Ge)) {
		t.Fatalf("expected equality < relational and LE == GE precedence")
	}
	// Additive vs Multiplicative
	if !(Precedence(Plus) == Precedence(Minus)) {
		t.Fatalf("expected Plus == Minus precedence")
	}
	if !(Precedence(Star) == Precedence(Slash) && Precedence(Slash) == Precedence(Percent)) {
		t.Fatalf("expected Star/Slash/Percent to share precedence")
	}
	if !(Precedence(Plus) < Precedence(Star)) {
		t.Fatalf("expected additive < multiplicative precedence")
	}
	// Non-operator returns zero
	if Precedence(Ident) != 0 || Precedence(KwFunc) != 0 {
		t.Fatalf("expected non-operators to have precedence 0")
	}
}

// TestKeywords_AllEntries checks that every declared keyword maps correctly and
// that a non-keyword returns Ident,false.
func testKeywords_AllEntries(t *testing.T) {
	for lex, kind := range Keywords {
		got, ok := LookupKeyword(lex)
		if !ok || got != kind {
			t.Fatalf("LookupKeyword(%q) => (%v,%v); want (%v,true)", lex, got, ok, kind)
		}
	}
	if got, ok := LookupKeyword("not_a_keyword"); ok || got != Ident {
		t.Fatalf("LookupKeyword(non-keyword) => (%v,%v); want (Ident,false)", got, ok)
	}
}

// TestOperators_AllEntries touches all operator lexemes in the table.
func testOperators_AllEntries(t *testing.T) {
	for lex, kind := range Operators {
		if kind == Unknown {
			t.Fatalf("operator %q mapped to Unknown", lex)
		}
		// basic sanity: every mapped operator should have non-zero precedence
		// except assignment and arrow which are not in precedence table yet.
		if kind != Assign && kind != Arrow && kind != Bang && Precedence(kind) == 0 {
			t.Fatalf("operator %q has zero precedence unexpectedly", lex)
		}
	}
	if _, ok := Operators["??"]; ok {
		t.Fatalf("unexpected operator mapping for ??")
	}
}
