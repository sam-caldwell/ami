package token

type Token struct {
	Kind   Kind
	Lexeme string
	Line   int
	Column int
	Offset int
}
