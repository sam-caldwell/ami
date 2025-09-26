package scanner

import "unicode/utf8"

// peekRune decodes and returns the next rune without advancing the scanner.
func (s *Scanner) peekRune() rune {
    if s.off+1 > len(s.src)-1 {
        return 0
    }
    r, _ := utf8.DecodeRuneInString(s.src[s.off+1:])
    return r
}
