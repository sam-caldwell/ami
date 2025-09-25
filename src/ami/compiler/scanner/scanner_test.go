package scanner

import (
    "testing"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func collectKinds(t *testing.T, s *Scanner) []tok.Token {
    t.Helper()
    var out []tok.Token
    for {
        tt := s.Next()
        out = append(out, tt)
        if tt.Kind == tok.EOF { break }
        if len(out) > 1024 { t.Fatalf("scanner runaway") }
    }
    return out
}

func TestScanner_KeywordsAndIdentifiers(t *testing.T) {
    s := New("package x import ingress transform collect egress pipeline state true false nil foo _bar")
    toks := collectKinds(t, s)
    kinds := []tok.Kind{
        tok.KW_PACKAGE, tok.IDENT,
        tok.KW_IMPORT,
        tok.KW_INGRESS, tok.KW_TRANSFORM, tok.KW_COLLECT, tok.KW_EGRESS,
        tok.KW_PIPELINE, tok.KW_STATE, tok.KW_TRUE, tok.KW_FALSE, tok.KW_NIL,
        tok.IDENT, tok.IDENT,
        tok.EOF,
    }
    if len(toks) != len(kinds) { t.Fatalf("len=%d want %d", len(toks), len(kinds)) }
    for i, k := range kinds {
        if toks[i].Kind != k {
            t.Fatalf("tok[%d].Kind=%v want %v (lexeme=%q)", i, toks[i].Kind, k, toks[i].Lexeme)
        }
    }
}

func TestScanner_OperatorsAndDelimiters(t *testing.T) {
    src := "( ) { } [ ] , ; : . = -> | == != < <= > >= + - * / %"
    s := New(src)
    want := []tok.Kind{
        tok.LPAREN, tok.RPAREN, tok.LBRACE, tok.RBRACE, tok.LBRACK, tok.RBRACK,
        tok.COMMA, tok.SEMI, tok.COLON, tok.DOT,
        tok.ASSIGN, tok.ARROW, tok.PIPE, tok.EQ, tok.NEQ,
        tok.LT, tok.LTE, tok.GT, tok.GTE,
        tok.PLUS, tok.MINUS, tok.STAR, tok.SLASH, tok.PERCENT,
        tok.EOF,
    }
    toks := collectKinds(t, s)
    if len(toks) != len(want) { t.Fatalf("len=%d want %d", len(toks), len(want)) }
    for i, k := range want {
        if toks[i].Kind != k {
            t.Fatalf("tok[%d].Kind=%v want %v (lexeme=%q)", i, toks[i].Kind, k, toks[i].Lexeme)
        }
    }
}

func TestScanner_CommentsAndWhitespaceSkipped(t *testing.T) {
    src := "// line comment\n/* block \n comment */ package /* mid */ main // end\n"
    toks := collectKinds(t, New(src))
    if len(toks) < 2 { t.Fatalf("expected tokens") }
    if toks[0].Kind != tok.KW_PACKAGE { t.Fatalf("first token kind=%v want KW_PACKAGE", toks[0].Kind) }
    if toks[1].Kind != tok.IDENT || toks[1].Lexeme != "main" { t.Fatalf("second token=%v want IDENT(main)", toks[1]) }
}

func TestScanner_StringsAndNumbers(t *testing.T) {
    s := New(`"a\"b" 123 45.67`)
    toks := collectKinds(t, s)
    want := []tok.Kind{tok.STRING, tok.NUMBER, tok.NUMBER, tok.EOF}
    if len(toks) != len(want) { t.Fatalf("len=%d want %d", len(toks), len(want)) }
    for i, k := range want { if toks[i].Kind != k { t.Fatalf("i=%d kind=%v want %v", i, toks[i].Kind, k) } }
}

func TestScanner_IllegalToken(t *testing.T) {
    s := New("@"); tt := s.Next()
    if tt.Kind != tok.ILLEGAL { t.Fatalf("kind=%v want ILLEGAL", tt.Kind) }
}
