package scanner

import (
    "strings"
    "unicode"
    "unicode/utf8"

    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

type Scanner struct {
    src    string
    off    int
    line   int
    column int
    pending []Comment
}

func New(src string) *Scanner { return &Scanner{src: src, line: 1, column: 1} }

// Comment captures a source comment with its starting position.
type Comment struct {
    Text   string
    Line   int
    Column int
    Offset int
}

func (s *Scanner) Next() tok.Token {
    s.skipSpaceAndComments()
    if s.off >= len(s.src) {
        return tok.Token{Kind: tok.EOF, Line: s.line, Column: s.column, Offset: s.off}
    }
    // Compiler directives: #pragma <directive> [payload...] to end of line
    if strings.HasPrefix(s.src[s.off:], "#pragma") {
        startCol := s.column
        startOff := s.off
        // consume '#pragma'
        s.off += len("#pragma")
        s.column += len("#pragma")
        // consume optional single space after pragma
        if s.off < len(s.src) {
            if s.src[s.off] == ' ' || s.src[s.off] == '\t' { s.off++; s.column++ }
        }
        // capture until newline or EOF
        start := s.off
        for s.off < len(s.src) {
            if s.src[s.off] == '\n' { break }
            s.off++
            s.column++
        }
        lex := strings.TrimSpace(s.src[start:s.off])
        // consume newline if present
        if s.off < len(s.src) && s.src[s.off] == '\n' {
            s.off++
            s.line++
            s.column = 1
        }
        return tok.Token{Kind: tok.PRAGMA, Lexeme: lex, Line: s.line - 1, Column: startCol, Offset: startOff}
    }
    r, w := utf8.DecodeRuneInString(s.src[s.off:])
    startCol := s.column
    startOff := s.off

    // Identifiers / keywords
    if unicode.IsLetter(r) || r == '_' {
        start := s.off
        s.advance(w)
        for s.off < len(s.src) {
            r, w = utf8.DecodeRuneInString(s.src[s.off:])
            if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
                break
            }
            s.advance(w)
        }
        lit := s.src[start:s.off]
        if kw, ok := tok.Keywords[strings.ToLower(lit)]; ok {
            return tok.Token{Kind: kw, Lexeme: lit, Line: s.line, Column: startCol, Offset: startOff}
        }
        return tok.Token{Kind: tok.IDENT, Lexeme: lit, Line: s.line, Column: startCol, Offset: startOff}
    }

    // Numbers (simple decimal / float)
    if unicode.IsDigit(r) {
        start := s.off
        s.advance(w)
        dotUsed := false
        for s.off < len(s.src) {
            r, w = utf8.DecodeRuneInString(s.src[s.off:])
            if r == '.' && !dotUsed {
                dotUsed = true
                s.advance(w)
                continue
            }
            if !unicode.IsDigit(r) {
                break
            }
            s.advance(w)
        }
        return tok.Token{Kind: tok.NUMBER, Lexeme: s.src[start:s.off], Line: s.line, Column: startCol, Offset: startOff}
    }

    // Strings (double quotes, naive escapes)
    if r == '"' {
        start := s.off
        s.advance(w)
        for s.off < len(s.src) {
            r, w = utf8.DecodeRuneInString(s.src[s.off:])
            if r == '\\' { // skip escape and next
                s.advance(w)
                if s.off < len(s.src) {
                    _, w2 := utf8.DecodeRuneInString(s.src[s.off:])
                    s.advance(w2)
                }
                continue
            }
            if r == '"' {
                s.advance(w)
                break
            }
            if r == '\n' {
                s.line++
                s.column = 1
                s.advance(w)
                continue
            }
            s.advance(w)
        }
        return tok.Token{Kind: tok.STRING, Lexeme: s.src[start:s.off], Line: s.line, Column: startCol, Offset: startOff}
    }

    // Multi-char operators
    // ->, ==, !=, <=, >=
    if r == '-' && s.peekRune() == '>' {
        s.advance(w)
        s.advance(1)
        return tok.Token{Kind: tok.ARROW, Lexeme: "->", Line: s.line, Column: startCol, Offset: startOff}
    }
    if r == '=' && s.peekRune() == '=' {
        s.advance(w)
        s.advance(1)
        return tok.Token{Kind: tok.EQ, Lexeme: "==", Line: s.line, Column: startCol, Offset: startOff}
    }
    if r == '!' && s.peekRune() == '=' {
        s.advance(w)
        s.advance(1)
        return tok.Token{Kind: tok.NEQ, Lexeme: "!=", Line: s.line, Column: startCol, Offset: startOff}
    }
    if r == '<' && s.peekRune() == '=' {
        s.advance(w)
        s.advance(1)
        return tok.Token{Kind: tok.LTE, Lexeme: "<=", Line: s.line, Column: startCol, Offset: startOff}
    }
    if r == '>' && s.peekRune() == '=' {
        s.advance(w)
        s.advance(1)
        return tok.Token{Kind: tok.GTE, Lexeme: ">=", Line: s.line, Column: startCol, Offset: startOff}
    }

    // Single-char tokens and fallback
    s.advance(w)
    switch r {
    case '(':
        return tok.Token{Kind: tok.LPAREN, Lexeme: "(", Line: s.line, Column: startCol, Offset: startOff}
    case ')':
        return tok.Token{Kind: tok.RPAREN, Lexeme: ")", Line: s.line, Column: startCol, Offset: startOff}
    case '{':
        return tok.Token{Kind: tok.LBRACE, Lexeme: "{", Line: s.line, Column: startCol, Offset: startOff}
    case '}':
        return tok.Token{Kind: tok.RBRACE, Lexeme: "}", Line: s.line, Column: startCol, Offset: startOff}
    case '[':
        return tok.Token{Kind: tok.LBRACK, Lexeme: "[", Line: s.line, Column: startCol, Offset: startOff}
    case ']':
        return tok.Token{Kind: tok.RBRACK, Lexeme: "]", Line: s.line, Column: startCol, Offset: startOff}
    case ',':
        return tok.Token{Kind: tok.COMMA, Lexeme: ",", Line: s.line, Column: startCol, Offset: startOff}
    case ';':
        return tok.Token{Kind: tok.SEMI, Lexeme: ";", Line: s.line, Column: startCol, Offset: startOff}
    case ':':
        return tok.Token{Kind: tok.COLON, Lexeme: ":", Line: s.line, Column: startCol, Offset: startOff}
    case '.':
        return tok.Token{Kind: tok.DOT, Lexeme: ".", Line: s.line, Column: startCol, Offset: startOff}
    case '=':
        return tok.Token{Kind: tok.ASSIGN, Lexeme: "=", Line: s.line, Column: startCol, Offset: startOff}
    case '|':
        return tok.Token{Kind: tok.PIPE, Lexeme: "|", Line: s.line, Column: startCol, Offset: startOff}
    case '+':
        return tok.Token{Kind: tok.PLUS, Lexeme: "+", Line: s.line, Column: startCol, Offset: startOff}
    case '-':
        return tok.Token{Kind: tok.MINUS, Lexeme: "-", Line: s.line, Column: startCol, Offset: startOff}
    case '*':
        return tok.Token{Kind: tok.STAR, Lexeme: "*", Line: s.line, Column: startCol, Offset: startOff}
    case '/':
        return tok.Token{Kind: tok.SLASH, Lexeme: "/", Line: s.line, Column: startCol, Offset: startOff}
    case '%':
        return tok.Token{Kind: tok.PERCENT, Lexeme: "%", Line: s.line, Column: startCol, Offset: startOff}
    case '&':
        return tok.Token{Kind: tok.AMP, Lexeme: "&", Line: s.line, Column: startCol, Offset: startOff}
    case '<':
        return tok.Token{Kind: tok.LT, Lexeme: "<", Line: s.line, Column: startCol, Offset: startOff}
    case '>':
        return tok.Token{Kind: tok.GT, Lexeme: ">", Line: s.line, Column: startCol, Offset: startOff}
    case '\n':
        s.line++
        s.column = 1
        return s.Next()
    default:
        return tok.Token{Kind: tok.ILLEGAL, Lexeme: string(r), Line: s.line, Column: startCol, Offset: startOff}
    }
}

func (s *Scanner) skipSpaceAndComments() {
    for s.off < len(s.src) {
        r, w := utf8.DecodeRuneInString(s.src[s.off:])
        // whitespace
        if unicode.IsSpace(r) {
            if r == '\n' {
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
            if idx := strings.IndexByte(s.src[s.off:], '\n'); idx >= 0 {
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
                if nl := strings.Count(segment, "\n"); nl > 0 {
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

// ConsumeComments returns and clears any comments encountered immediately
// before the next non-space token.
func (s *Scanner) ConsumeComments() []Comment {
    if len(s.pending) == 0 { return nil }
    out := make([]Comment, len(s.pending))
    copy(out, s.pending)
    s.pending = nil
    return out
}

func (s *Scanner) peekRune() rune {
    if s.off+1 > len(s.src)-1 {
        return 0
    }
    r, _ := utf8.DecodeRuneInString(s.src[s.off+1:])
    return r
}

func (s *Scanner) advance(w int) {
    s.off += w
    s.column += w
}
