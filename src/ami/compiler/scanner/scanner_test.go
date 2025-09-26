package scanner

import (
    "strings"
    "testing"

    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func collectKinds(t *testing.T, s *Scanner) []tok.Token {
    t.Helper()
    var out []tok.Token
    for {
        tt := s.Next()
        out = append(out, tt)
        if tt.Kind == tok.EOF {
            break
        }
        if len(out) > 1024 {
            t.Fatalf("scanner runaway")
        }
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
    if len(toks) != len(kinds) {
        t.Fatalf("len=%d want %d", len(toks), len(kinds))
    }
    for i, k := range kinds {
        if toks[i].Kind != k {
            t.Fatalf("tok[%d].Kind=%v want %v (lexeme=%q)", i, toks[i].Kind, k, toks[i].Lexeme)
        }
    }
}

func TestScanner_OperatorsAndDelimiters(t *testing.T) {
    src := "( ) { } [ ] , ; : . = -> | == != < <= > >= + - * / % &"
    s := New(src)
    want := []tok.Kind{
        tok.LPAREN, tok.RPAREN, tok.LBRACE, tok.RBRACE, tok.LBRACK, tok.RBRACK,
        tok.COMMA, tok.SEMI, tok.COLON, tok.DOT,
        tok.ASSIGN, tok.ARROW, tok.PIPE, tok.EQ, tok.NEQ,
        tok.LT, tok.LTE, tok.GT, tok.GTE,
        tok.PLUS, tok.MINUS, tok.STAR, tok.SLASH, tok.PERCENT, tok.AMP,
        tok.EOF,
    }
    toks := collectKinds(t, s)
    if len(toks) != len(want) {
        t.Fatalf("len=%d want %d", len(toks), len(want))
    }
    for i, k := range want {
        if toks[i].Kind != k {
            t.Fatalf("tok[%d].Kind=%v want %v (lexeme=%q)", i, toks[i].Kind, k, toks[i].Lexeme)
        }
    }
}

func TestScanner_CommentsAndWhitespaceSkipped(t *testing.T) {
    src := "// line comment\n/* block \n comment */ package /* mid */ main // end\n"
    toks := collectKinds(t, New(src))
    if len(toks) < 2 {
        t.Fatalf("expected tokens")
    }
    if toks[0].Kind != tok.KW_PACKAGE {
        t.Fatalf("first token kind=%v want KW_PACKAGE", toks[0].Kind)
    }
    if toks[1].Kind != tok.IDENT || toks[1].Lexeme != "main" {
        t.Fatalf("second token=%v want IDENT(main)", toks[1])
    }
}

func TestScanner_StringsAndNumbers(t *testing.T) {
    s := New(`"a\"b" 123 45.67`)
    toks := collectKinds(t, s)
    want := []tok.Kind{tok.STRING, tok.NUMBER, tok.NUMBER, tok.EOF}
    if len(toks) != len(want) {
        t.Fatalf("len=%d want %d", len(toks), len(want))
    }
    for i, k := range want {
        if toks[i].Kind != k {
            t.Fatalf("i=%d kind=%v want %v", i, toks[i].Kind, k)
        }
    }
}

func TestScanner_IllegalToken(t *testing.T) {
    s := New("@")
    tt := s.Next()
    if tt.Kind != tok.ILLEGAL {
        t.Fatalf("kind=%v want ILLEGAL", tt.Kind)
    }
}

func TestScanner_Pragma_Directives(t *testing.T) {
    src := "#pragma capabilities net,fs\n#pragma trust sandboxed\n#pragma backpressure drop\npackage x"
    s := New(src)
    // Expect three pragma tokens then package
    t1 := s.Next()
    t2 := s.Next()
    t3 := s.Next()
    t4 := s.Next()
    if t1.Kind != tok.PRAGMA || t1.Lexeme == "" {
        t.Fatalf("t1 not PRAGMA: %+v", t1)
    }
    if t2.Kind != tok.PRAGMA || t3.Kind != tok.PRAGMA {
        t.Fatalf("pragma sequence missing")
    }
    if got, want := strings.Fields(t1.Lexeme)[0], "capabilities"; got != want {
        t.Fatalf("want first pragma 'capabilities', got %q", got)
    }
    if got, want := strings.Fields(t2.Lexeme)[0], "trust"; got != want {
        t.Fatalf("want second pragma 'trust', got %q", got)
    }
    if got, want := strings.Fields(t3.Lexeme)[0], "backpressure"; got != want {
        t.Fatalf("want third pragma 'backpressure', got %q", got)
    }
    if t4.Kind != tok.KW_PACKAGE {
        t.Fatalf("expected package after pragmas, got %v", t4.Kind)
    }
}

