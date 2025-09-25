package token

type Kind int

const (
    EOF Kind = iota
    IDENT
    NUMBER
    STRING
    ILLEGAL
)

type Token struct {
    Kind   Kind
    Lexeme string
    Line   int
    Column int
}
