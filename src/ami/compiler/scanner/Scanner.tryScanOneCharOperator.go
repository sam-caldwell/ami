package scanner

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// tryScanOneCharOperator scans a single-character operator if present.
func (s *Scanner) tryScanOneCharOperator(src string, n int, start int) (token.Token, bool) {
    if k, ok := token.Operators[string(src[s.offset])]; ok {
        s.offset++
        return token.Token{Kind: k, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}, true
    }
    return token.Token{}, false
}

