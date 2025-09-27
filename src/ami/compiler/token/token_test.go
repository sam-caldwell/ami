package token

import "testing"

func TestToken_Basics(t *testing.T) {
    tok := Token{Kind: Ident, Lexeme: "name"}
    if tok.Kind != Ident || tok.Lexeme != "name" {
        t.Fatalf("unexpected token: %+v", tok)
    }
}

