package scanner

import (
	"strings"
	"unicode"
	"unicode/utf8"

	tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// Next returns the next token from the source, skipping leading whitespace
// and comments. It also recognizes compiler directives beginning with
// "#pragma" and returns them as PRAGMA tokens with payload in Lexeme.
func (s *Scanner) Next() tok.Token {
	s.skipSpaceAndComments()
	if s.off >= len(s.src) {
		return tok.Token{Kind: tok.EOF, Line: s.line, Column: s.column, Offset: s.off}
	}
	// Compiler directives: #pragma <directive> [payload...] to end of line
	if strings.HasPrefix(s.src[s.off:], tok.LexPragma) {
		startLine := s.line
		startCol := s.column
		startOff := s.off
		// consume '#pragma'
		s.off += len(tok.LexPragma)
		s.column += len(tok.LexPragma)
		// consume optional single space after pragma
		consumeSpace(s)
		// capture until newline or EOF
		start := s.off
		captureUntilEndOfLine(s)
		lex := strings.TrimSpace(s.src[start:s.off])
		consumeNewline(s)
		return tok.Token{
			Kind:   tok.PRAGMA,
			Lexeme: lex,
			Line:   startLine,
			Column: startCol,
			Offset: startOff,
		}
	}
	r, w := utf8.DecodeRuneInString(s.src[s.off:])
	startCol := s.column
	startOff := s.off

	identifierEnd := func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == tok.LexUnderscore)
	}

	// Identifiers / keywords
	if unicode.IsLetter(r) || r == tok.LexUnderscore {
		start := s.off
		s.advance(w)
		for s.off < len(s.src) {
			r, w = utf8.DecodeRuneInString(s.src[s.off:])
			if identifierEnd(r) {
				break
			}
			s.advance(w)
		}
		lit := s.src[start:s.off]
		if kw, ok := tok.Keywords[strings.ToLower(lit)]; ok {
			return tok.Token{
				Kind:   kw,
				Lexeme: lit,
				Line:   s.line,
				Column: startCol,
				Offset: startOff,
			}
		}
		return tok.Token{
			Kind:   tok.IDENT,
			Lexeme: lit,
			Line:   s.line,
			Column: startCol,
			Offset: startOff,
		}
	}

	// Numbers (simple decimal / float)
	if unicode.IsDigit(r) {
		start := s.off
		s.advance(w)
		dotUsed := false
		for s.off < len(s.src) {
			r, w = utf8.DecodeRuneInString(s.src[s.off:])
			if r == tok.LexPeriod && !dotUsed {
				dotUsed = true
				s.advance(w)
				continue
			}
			if !unicode.IsDigit(r) {
				break
			}
			s.advance(w)
		}
		return tok.Token{
			Kind:   tok.NUMBER,
			Lexeme: s.src[start:s.off],
			Line:   s.line,
			Column: startCol,
			Offset: startOff,
		}
	}

	// Strings (double quotes, naive escapes)
	if r == tok.LexDblQuote {
		start := s.off
		s.advance(w)
		for s.off < len(s.src) {
			r, w = utf8.DecodeRuneInString(s.src[s.off:])
			if r == tok.LexBkSlash {
				s.advance(w)
				if s.off < len(s.src) {
					_, w2 := utf8.DecodeRuneInString(s.src[s.off:])
					s.advance(w2)
				}
				continue
			}
			if r == tok.LexDblQuote {
				s.advance(w)
				break
			}
			if r == tok.LexLf {
				s.line++
				s.column = 1
				s.advance(w)
				continue
			}
			s.advance(w)
		}
		return tok.Token{
			Kind:   tok.STRING,
			Lexeme: s.src[start:s.off],
			Line:   s.line,
			Column: startCol,
			Offset: startOff,
		}
	}

	// Multi-char operators
	// ->, ==, !=, <=, >=
	if r == tok.LexHyphen && s.peekRune() == tok.LexGt {
		s.advance(w)
		s.advance(1)
		return tok.Token{
			Kind:   tok.ARROW,
			Lexeme: tok.LexArrowRight,
			Line:   s.line,
			Column: startCol,
			Offset: startOff,
		}
	}
	if r == tok.LexEQ && s.peekRune() == tok.LexEQ {
		s.advance(w)
		s.advance(1)
		return tok.Token{
			Kind:   tok.EQ,
			Lexeme: tok.LexBoolEQ,
			Line:   s.line,
			Column: startCol,
			Offset: startOff,
		}
	}
	if r == tok.LexExclamation && s.peekRune() == tok.LexEQ {
		s.advance(w)
		s.advance(1)
		return tok.Token{
			Kind:   tok.NEQ,
			Lexeme: tok.LexBoolNE,
			Line:   s.line,
			Column: startCol,
			Offset: startOff,
		}
	}
	if r == tok.LexLt && s.peekRune() == tok.LexEQ {
		s.advance(w)
		s.advance(1)
		return tok.Token{
			Kind:   tok.LTE,
			Lexeme: tok.LexBoolLE,
			Line:   s.line,
			Column: startCol,
			Offset: startOff,
		}
	}
	if r == tok.LexGt && s.peekRune() == tok.LexEQ {
		s.advance(w)
		s.advance(1)
		return tok.Token{
			Kind:   tok.GTE,
			Lexeme: tok.LexBoolGE,
			Line:   s.line,
			Column: startCol,
			Offset: startOff,
		}
	}

	// Single-char tokens and fallback
	s.advance(w)
	switch r {
	case tok.LexLParen:
		return tok.Token{Kind: tok.LPAREN, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexRParen:
		return tok.Token{Kind: tok.RPAREN, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexLBrace:
		return tok.Token{Kind: tok.LBRACE, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexRBrace:
		return tok.Token{Kind: tok.RBRACE, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexLBracket:
		return tok.Token{Kind: tok.LBRACK, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexRBracket:
		return tok.Token{Kind: tok.RBRACK, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexComma:
		return tok.Token{Kind: tok.COMMA, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexSemicolon:
		return tok.Token{Kind: tok.SEMI, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexColon:
		return tok.Token{Kind: tok.COLON, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexPeriod:
		return tok.Token{Kind: tok.DOT, Lexeme: string(tok.LexPeriod), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexEQ:
		return tok.Token{Kind: tok.ASSIGN, Lexeme: string(tok.LexEQ), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexPipe:
		return tok.Token{Kind: tok.PIPE, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexPlus:
		return tok.Token{Kind: tok.PLUS, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexHyphen:
		return tok.Token{Kind: tok.MINUS, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexAsterisk:
		return tok.Token{Kind: tok.STAR, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexSlash:
		return tok.Token{Kind: tok.SLASH, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexPercent:
		return tok.Token{Kind: tok.PERCENT, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexAmpersand:
		return tok.Token{Kind: tok.AMP, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexLt:
		return tok.Token{Kind: tok.LT, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexGt:
		return tok.Token{Kind: tok.GT, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	case tok.LexLf:
		s.line++
		s.column = 1
		return s.Next()
	default:
		return tok.Token{Kind: tok.ILLEGAL, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
	}
}
