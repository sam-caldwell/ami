package scanner

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// tryScanPunctuationSymbol scans punctuation symbols mapped to distinct kinds.
func (s *Scanner) tryScanPunctuationSymbol(src string, n int, start int) (token.Token, bool) {
    switch src[s.offset] {
    case '(':
        s.offset++
        return token.Token{Kind: token.LParenSym, Lexeme: token.LParen, Pos: s.file.Pos(start)}, true
    case ')':
        s.offset++
        return token.Token{Kind: token.RParenSym, Lexeme: token.RParen, Pos: s.file.Pos(start)}, true
    case '{':
        s.offset++
        return token.Token{Kind: token.LBraceSym, Lexeme: token.LBrace, Pos: s.file.Pos(start)}, true
    case '}':
        s.offset++
        return token.Token{Kind: token.RBraceSym, Lexeme: token.RBrace, Pos: s.file.Pos(start)}, true
    case '[':
        s.offset++
        return token.Token{Kind: token.LBracketSym, Lexeme: "[", Pos: s.file.Pos(start)}, true
    case ']':
        s.offset++
        return token.Token{Kind: token.RBracketSym, Lexeme: "]", Pos: s.file.Pos(start)}, true
    case ',':
        s.offset++
        return token.Token{Kind: token.CommaSym, Lexeme: token.Comma, Pos: s.file.Pos(start)}, true
    case ';':
        s.offset++
        return token.Token{Kind: token.SemiSym, Lexeme: token.Semi, Pos: s.file.Pos(start)}, true
    case '.':
        s.offset++
        return token.Token{Kind: token.DotSym, Lexeme: ".", Pos: s.file.Pos(start)}, true
    case ':':
        s.offset++
        return token.Token{Kind: token.ColonSym, Lexeme: ":", Pos: s.file.Pos(start)}, true
    case '|':
        s.offset++
        return token.Token{Kind: token.PipeSym, Lexeme: "|", Pos: s.file.Pos(start)}, true
    case '\\':
        s.offset++
        return token.Token{Kind: token.BackslashSym, Lexeme: "\\", Pos: s.file.Pos(start)}, true
    case '$':
        s.offset++
        return token.Token{Kind: token.DollarSym, Lexeme: "$", Pos: s.file.Pos(start)}, true
    case '`':
        s.offset++
        return token.Token{Kind: token.TickSym, Lexeme: "`", Pos: s.file.Pos(start)}, true
    case '~':
        s.offset++
        return token.Token{Kind: token.TildeSym, Lexeme: "~", Pos: s.file.Pos(start)}, true
    case '?':
        s.offset++
        return token.Token{Kind: token.QuestionSym, Lexeme: "?", Pos: s.file.Pos(start)}, true
    case '@':
        s.offset++
        return token.Token{Kind: token.AtSym, Lexeme: "@", Pos: s.file.Pos(start)}, true
    case '#':
        s.offset++
        return token.Token{Kind: token.PoundSym, Lexeme: "#", Pos: s.file.Pos(start)}, true
    case '^':
        s.offset++
        return token.Token{Kind: token.CaretSym, Lexeme: "^", Pos: s.file.Pos(start)}, true
    case '\'':
        s.offset++
        return token.Token{Kind: token.SingleQuoteSym, Lexeme: "'", Pos: s.file.Pos(start)}, true
    case '"':
        s.offset++
        return token.Token{Kind: token.DoubleQuoteSym, Lexeme: "\"", Pos: s.file.Pos(start)}, true
    }
    return token.Token{}, false
}

