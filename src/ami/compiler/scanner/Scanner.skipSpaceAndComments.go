package scanner

import (
	"strings"
	"unicode"
	"unicode/utf8"

	tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (s *Scanner) skipSpaceAndComments() {
	for s.off < len(s.src) {
		r, w := utf8.DecodeRuneInString(s.src[s.off:])
		// whitespace
		if unicode.IsSpace(r) {
			if r == tok.LexLf {
				s.line++
				s.column = 1
			} else {
				s.column += w
			}
			s.off += w
			continue
		}
		// line comment //...
		if r == '/' && s.peekRune() == '/' {
			// record comment start
			cLine, cCol, cOff := s.line, s.column, s.off
			// skip '//'
			s.off += 2
			// find next newline
			if idx := strings.IndexByte(s.src[s.off:], tok.LexLf); idx >= 0 {
				// comment text excludes the trailing newline
				text := s.src[s.off : s.off+idx]
				s.pending = append(s.pending, Comment{Text: text, Line: cLine, Column: cCol, Offset: cOff})
				s.off += idx + 1
				s.line++
				s.column = 1
			} else {
				// no newline; consume to end
				text := s.src[s.off:]
				s.pending = append(s.pending, Comment{Text: text, Line: cLine, Column: cCol, Offset: cOff})
				s.off = len(s.src)
			}
			continue
		}
		// block comment /* ... */ (no nesting)
		if r == '/' && s.peekRune() == '*' {
			// record start
			cLine, cCol, cOff := s.line, s.column, s.off
			// skip '/*'
			s.off += 2
			// find closing '*/'
			if idx := strings.Index(s.src[s.off:], "*/"); idx >= 0 {
				segment := s.src[s.off : s.off+idx]
				s.pending = append(s.pending, Comment{Text: segment, Line: cLine, Column: cCol, Offset: cOff})
				// count newlines to adjust line/column
				if nl := strings.Count(segment, string(tok.LexLf)); nl > 0 {
					s.line += nl
					s.column = 1
				} else {
					s.column += len(segment)
				}
				s.off += idx + 2
			} else {
				// unterminated; consume to end
				seg := s.src[s.off:]
				s.pending = append(s.pending, Comment{Text: seg, Line: cLine, Column: cCol, Offset: cOff})
				s.off = len(s.src)
			}
			continue
		}
		break
	}
}
