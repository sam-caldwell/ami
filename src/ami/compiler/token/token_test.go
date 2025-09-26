package token

import "testing"

func TestToken_Struct_Fields(t *testing.T) {
    tok := Token{Kind: IDENT, Lexeme: "name", Line: 3, Column: 5, Offset: 12}
    if tok.Kind != IDENT || tok.Lexeme != "name" || tok.Line != 3 || tok.Column != 5 || tok.Offset != 12 {
        t.Fatalf("unexpected token: %+v", tok)
    }
}

