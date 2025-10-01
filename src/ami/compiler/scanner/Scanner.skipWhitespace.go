package scanner

import (
    "unicode"
    "unicode/utf8"
)

// skipWhitespace advances the scanner past any whitespace runes.
func (s *Scanner) skipWhitespace(src string, n int) {
    for s.offset < n {
        r, size := utf8.DecodeRuneInString(src[s.offset:])
        if r == utf8.RuneError && size == 1 { // invalid byte
            break
        }
        if !unicode.IsSpace(r) { break }
        s.offset += size
    }
}

