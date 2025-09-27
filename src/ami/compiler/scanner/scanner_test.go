package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_IdentAndEOF(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "package main"}
    s := New(f)
    t1 := s.Next()
    if t1.Kind != token.KwPackage || t1.Lexeme != "package" { t.Fatalf("unexpected t1: %+v", t1) }
    t2 := s.Next()
    if t2.Kind != token.Ident || t2.Lexeme != "main" { t.Fatalf("unexpected t2: %+v", t2) }
    t3 := s.Next()
    if t3.Kind != token.EOF { t.Fatalf("expected EOF, got: %+v", t3) }
}

func TestScanner_NilOrEmpty(t *testing.T) {
    var s *Scanner
    if tok := s.Next(); tok.Kind != token.EOF { t.Fatalf("nil scanner should return EOF") }
    s = New(&source.File{Name: "t", Content: ""})
    if tok := s.Next(); tok.Kind != token.EOF { t.Fatalf("empty file should return EOF") }
}

func TestScanner_CommentsAndSymbols(t *testing.T) {
    src := `// line comment
(/* block */) ,;`
    s := New(&source.File{Name: "t.ami", Content: src})
    // Expect line comment first
    t1 := s.Next()
    if t1.Kind != token.LineComment || t1.Lexeme == "" { t.Fatalf("want line comment, got %+v", t1) }
    // block comment next
    tbc := s.Next()
    if tbc.Kind != token.BlockComment || tbc.Lexeme == "" { t.Fatalf("want block comment, got %+v", tbc) }
    // then LParen symbol
    t1 = s.Next()
    if t1.Kind != token.LParenSym || t1.Lexeme != token.LParen { t.Fatalf("want LParen sym, got %+v", t1) }
    // RParen
    t2 := s.Next()
    if t2.Kind != token.RParenSym || t2.Lexeme != token.RParen { t.Fatalf("want RParen sym, got %+v", t2) }
    // comma
    t3 := s.Next()
    if t3.Kind != token.CommaSym || t3.Lexeme != token.Comma { t.Fatalf("want comma sym, got %+v", t3) }
    // semi
    t4 := s.Next()
    if t4.Kind != token.SemiSym || t4.Lexeme != token.Semi { t.Fatalf("want semicolon sym, got %+v", t4) }
    // EOF
    if tok := s.Next(); tok.Kind != token.EOF { t.Fatalf("expected EOF, got %+v", tok) }
}

func TestScanner_Numbers_Strings_Operators(t *testing.T) {
    src := `123 "hi" == != <= >= && || -> + - * / % !`
    s := New(&source.File{Name: "t", Content: src})
    // number
    if tok := s.Next(); tok.Kind != token.Number || tok.Lexeme != "123" { t.Fatalf("num: %+v", tok) }
    // string
    if tok := s.Next(); tok.Kind != token.String || tok.Lexeme != `"hi"` { t.Fatalf("str: %+v", tok) }
    // operators (2-char first)
    wantOps := []struct{ lex string; k token.Kind }{
        {"==", token.Eq}, {"!=", token.Ne}, {"<=", token.Le}, {">=", token.Ge},
        {"&&", token.And}, {"||", token.Or}, {"->", token.Arrow},
        {"+", token.Plus}, {"-", token.Minus}, {"*", token.Star}, {"/", token.Slash}, {"%", token.Percent}, {"!", token.Bang},
    }
    for _, w := range wantOps {
        tok := s.Next()
        if tok.Kind != w.k || tok.Lexeme != w.lex { t.Fatalf("op %q => %+v", w.lex, tok) }
    }
    if tok := s.Next(); tok.Kind != token.EOF { t.Fatalf("expected EOF, got %+v", tok) }
}

func TestScanner_UTF8_Identifiers_And_Numerics(t *testing.T) {
    src := "π = 3.14 0x1f 0b1010 0o77 1e9 2.5e-3"
    s := New(&source.File{Name: "t", Content: src})
    // π
    if tok := s.Next(); tok.Kind != token.Ident || tok.Lexeme != "π" { t.Fatalf("utf8 ident: %+v", tok) }
    // '='
    if tok := s.Next(); tok.Kind != token.Assign { t.Fatalf("assign: %+v", tok) }
    // decimals and variants
    want := []string{"3.14", "0x1f", "0b1010", "0o77", "1e9", "2.5e-3"}
    for _, w := range want {
        tok := s.Next()
        if tok.Kind != token.Number || tok.Lexeme != w { t.Fatalf("num %q => %+v", w, tok) }
    }
}

func TestScanner_StringEscapes_And_Unterminated(t *testing.T) {
    s := New(&source.File{Name: "t", Content: `"a\"b"`})
    tok := s.Next()
    if tok.Kind != token.String || tok.Lexeme != `"a\"b"` { t.Fatalf("escaped string: %+v", tok) }
    s = New(&source.File{Name: "t", Content: `"unterminated`})
    tok = s.Next()
    if tok.Kind != token.Unknown { t.Fatalf("unterminated should be Unknown: %+v", tok) }
}
