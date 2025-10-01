package scanner

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// tryScanTwoCharOperator scans a two-character operator if present.
func (s *Scanner) tryScanTwoCharOperator(src string, n int, start int) (token.Token, bool) {
    if s.offset+1 >= n { return token.Token{}, false }
    two := src[s.offset : s.offset+2]
    if k, ok := token.Operators[two]; ok {
        s.offset += 2
        return token.Token{Kind: k, Lexeme: two, Pos: s.file.Pos(start)}, true
    }
    return token.Token{}, false
}

