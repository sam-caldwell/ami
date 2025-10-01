package scanner

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// scanFallbackSymbol returns a generic Symbol token and advances one byte.
func (s *Scanner) scanFallbackSymbol(src string, n int, start int) token.Token {
    s.offset++
    return token.Token{Kind: token.Symbol, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
}

