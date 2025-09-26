package token

// Token is a lexical token with kind and source position.
type Token struct {
    Kind   Kind
    Lexeme string
    Line   int
    Column int
    Offset int
}
