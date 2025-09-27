package token

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Token represents a lexical token with its kind and literal text.
// Position information is provided by the scanner/source packages.
type Token struct {
    Kind   Kind
    Lexeme string
    Pos    source.Position
}
