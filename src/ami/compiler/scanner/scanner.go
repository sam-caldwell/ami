package scanner

import (
    "strings"
    "unicode"
    "unicode/utf8"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// Scanner performs a minimal lexical scan over a source.File.
type Scanner struct {
    file   *source.File
    offset int
}

// New creates a new Scanner for the provided file.
func New(f *source.File) *Scanner { return &Scanner{file: f, offset: 0} }

// Next returns the next token. Handles identifiers/keywords, numbers, strings,
// operators (including basic 2‑rune forms), symbols, and comments.
func (s *Scanner) Next() token.Token {
    if s == nil || s.file == nil {
        return token.Token{Kind: token.EOF, Pos: source.Position{}}
    }
    src := s.file.Content
    n := len(src)
    // skip whitespace (UTF‑8 aware)
    for s.offset < n {
        r, size := utf8.DecodeRuneInString(src[s.offset:])
        if r == utf8.RuneError && size == 1 { // invalid byte
            break
        }
        if !unicode.IsSpace(r) { break }
        s.offset += size
    }
    if s.offset >= n {
        return token.Token{Kind: token.EOF, Pos: s.file.Pos(s.offset)}
    }

    start := s.offset
    // comments as tokens
    if s.offset+1 < n && src[s.offset] == '/' {
        if src[s.offset+1] == '/' {
            start := s.offset
            s.offset += 2
            cstart := s.offset
            for s.offset < n {
                r, size := utf8.DecodeRuneInString(src[s.offset:])
                if r == '\n' || size == 0 { break }
                s.offset += size
            }
            text := src[cstart:s.offset]
            return token.Token{Kind: token.LineComment, Lexeme: text, Pos: s.file.Pos(start)}
        }
        if src[s.offset+1] == '*' {
            start := s.offset
            s.offset += 2
            cstart := s.offset
            for s.offset+1 < n && !(src[s.offset] == '*' && src[s.offset+1] == '/') { s.offset++ }
            text := src[cstart:s.offset]
            if s.offset+1 < n { s.offset += 2 }
            return token.Token{Kind: token.BlockComment, Lexeme: text, Pos: s.file.Pos(start)}
        }
    }

    ch, size := utf8.DecodeRuneInString(src[s.offset:])

    // identifier or keyword: [A-Za-z_][A-Za-z0-9_]* (underscore allowed)
    if ch == '_' || unicode.IsLetter(ch) {
        s.offset += size
        for s.offset < n {
            r, sz := utf8.DecodeRuneInString(src[s.offset:])
            if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
                s.offset += sz
                continue
            }
            break
        }
        lex := src[start:s.offset]
        if k, ok := token.LookupKeyword(strings.ToLower(lex)); ok {
            return token.Token{Kind: k, Lexeme: lex, Pos: s.file.Pos(start)}
        }
        return token.Token{Kind: token.Ident, Lexeme: lex, Pos: s.file.Pos(start)}
    }

    // number: simple digits sequence
    if unicode.IsDigit(ch) {
        // radix prefixes: 0x, 0b, 0o
        if src[s.offset] == '0' && s.offset+1 < n {
            switch src[s.offset+1] {
            case 'x', 'X':
                s.offset += 2
                for s.offset < n {
                    r, sz := utf8.DecodeRuneInString(src[s.offset:])
                    if ('0' <= r && r <= '9') || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F') { s.offset += sz; continue }
                    break
                }
                return token.Token{Kind: token.Number, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
            case 'b', 'B':
                s.offset += 2
                for s.offset < n {
                    r, sz := utf8.DecodeRuneInString(src[s.offset:])
                    if r == '0' || r == '1' { s.offset += sz; continue }
                    break
                }
                return token.Token{Kind: token.Number, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
            case 'o', 'O':
                s.offset += 2
                for s.offset < n {
                    r, sz := utf8.DecodeRuneInString(src[s.offset:])
                    if '0' <= r && r <= '7' { s.offset += sz; continue }
                    break
                }
                return token.Token{Kind: token.Number, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
            }
        }
        // decimal int/float
        s.offset += size
        for s.offset < n {
            r, sz := utf8.DecodeRuneInString(src[s.offset:])
            if unicode.IsDigit(r) { s.offset += sz; continue }
            break
        }
        // fractional
        if s.offset < n && src[s.offset] == '.' {
            s.offset++
            for s.offset < n {
                r, sz := utf8.DecodeRuneInString(src[s.offset:])
                if unicode.IsDigit(r) { s.offset += sz; continue }
                break
            }
        }
        // exponent
        if s.offset < n && (src[s.offset] == 'e' || src[s.offset] == 'E') {
            s.offset++
            if s.offset < n && (src[s.offset] == '+' || src[s.offset] == '-') { s.offset++ }
            for s.offset < n {
                r, sz := utf8.DecodeRuneInString(src[s.offset:])
                if unicode.IsDigit(r) { s.offset += sz; continue }
                break
            }
        }
        return token.Token{Kind: token.Number, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
    }

    // string literal: "..." (no escape handling; minimal)
    if ch == '"' {
        s.offset += size // skip opening quote
        for s.offset < n {
            r, sz := utf8.DecodeRuneInString(src[s.offset:])
            if r == '"' { s.offset += sz; return token.Token{Kind: token.String, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)} }
            if r == '\\' {
                // consume escape sequence: one rune following
                s.offset += sz
                if s.offset < n {
                    _, esz := utf8.DecodeRuneInString(src[s.offset:])
                    s.offset += esz
                    continue
                }
                break
            }
            s.offset += sz
        }
        // Unterminated string → Unknown token
        return token.Token{Kind: token.Unknown, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
    }

    // operators: prefer 2‑char forms
    if s.offset+1 < n {
        two := src[s.offset : s.offset+2]
        if k, ok := token.Operators[two]; ok {
            s.offset += 2
            return token.Token{Kind: k, Lexeme: two, Pos: s.file.Pos(start)}
        }
    }
    // single char operator
    if k, ok := token.Operators[string(src[s.offset])]; ok {
        s.offset++
        return token.Token{Kind: k, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
    }

    // punctuation mapping to distinct kinds
    switch src[s.offset] {
    case '(':
        s.offset++
        return token.Token{Kind: token.LParenSym, Lexeme: token.LParen, Pos: s.file.Pos(start)}
    case ')':
        s.offset++
        return token.Token{Kind: token.RParenSym, Lexeme: token.RParen, Pos: s.file.Pos(start)}
    case '{':
        s.offset++
        return token.Token{Kind: token.LBraceSym, Lexeme: token.LBrace, Pos: s.file.Pos(start)}
    case '}':
        s.offset++
        return token.Token{Kind: token.RBraceSym, Lexeme: token.RBrace, Pos: s.file.Pos(start)}
    case '[':
        s.offset++
        return token.Token{Kind: token.LBracketSym, Lexeme: "[", Pos: s.file.Pos(start)}
    case ']':
        s.offset++
        return token.Token{Kind: token.RBracketSym, Lexeme: "]", Pos: s.file.Pos(start)}
    case ',':
        s.offset++
        return token.Token{Kind: token.CommaSym, Lexeme: token.Comma, Pos: s.file.Pos(start)}
    case ';':
        s.offset++
        return token.Token{Kind: token.SemiSym, Lexeme: token.Semi, Pos: s.file.Pos(start)}
    case '.':
        s.offset++
        return token.Token{Kind: token.DotSym, Lexeme: ".", Pos: s.file.Pos(start)}
    case ':':
        s.offset++
        return token.Token{Kind: token.ColonSym, Lexeme: ":", Pos: s.file.Pos(start)}
    case '|':
        s.offset++
        return token.Token{Kind: token.PipeSym, Lexeme: "|", Pos: s.file.Pos(start)}
    case '\\':
        s.offset++
        return token.Token{Kind: token.BackslashSym, Lexeme: "\\", Pos: s.file.Pos(start)}
    case '$':
        s.offset++
        return token.Token{Kind: token.DollarSym, Lexeme: "$", Pos: s.file.Pos(start)}
    case '`':
        s.offset++
        return token.Token{Kind: token.TickSym, Lexeme: "`", Pos: s.file.Pos(start)}
    case '~':
        s.offset++
        return token.Token{Kind: token.TildeSym, Lexeme: "~", Pos: s.file.Pos(start)}
    case '?':
        s.offset++
        return token.Token{Kind: token.QuestionSym, Lexeme: "?", Pos: s.file.Pos(start)}
    case '@':
        s.offset++
        return token.Token{Kind: token.AtSym, Lexeme: "@", Pos: s.file.Pos(start)}
    case '#':
        s.offset++
        return token.Token{Kind: token.PoundSym, Lexeme: "#", Pos: s.file.Pos(start)}
    case '^':
        s.offset++
        return token.Token{Kind: token.CaretSym, Lexeme: "^", Pos: s.file.Pos(start)}
    case '\'':
        s.offset++
        return token.Token{Kind: token.SingleQuoteSym, Lexeme: "'", Pos: s.file.Pos(start)}
    case '"':
        // handled by string literal logic above; but if reached here treat as symbol
        s.offset++
        return token.Token{Kind: token.DoubleQuoteSym, Lexeme: "\"", Pos: s.file.Pos(start)}
    }

    // fallback: single unknown symbol
    s.offset++
    return token.Token{Kind: token.Symbol, Lexeme: src[start:s.offset], Pos: s.file.Pos(start)}
}
