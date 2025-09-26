package parser

import (
    "testing"

    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// TestDebug_ParseBodyStmts isolates body parsing on captured tokens.
func TestDebug_ParseBodyStmts(t *testing.T) {
    src := `package p
func f(a int) (int,int) {
  // before var
  var x = a + 1
  // before expr
  a + 2
  // before assign
  *a = a + 3
  // before defer
  defer f(a)
  // before return
  return a, a
}`
    p := New(src)
    // replicate body capture logic roughly
    // advance to '{'
    for p.cur.Kind != tok.LBRACE && p.cur.Kind != tok.EOF {
        p.next()
    }
    if p.cur.Kind != tok.LBRACE {
        t.Fatalf("no body start")
    }
    depth := 1
    p.next()
    var body []tok.Token
    for depth > 0 && p.cur.Kind != tok.EOF {
        if p.cur.Kind == tok.LBRACE {
            depth++
        }
        if p.cur.Kind == tok.RBRACE {
            depth--
            if depth == 0 {
                p.next()
                break
            }
        }
        body = append(body, p.cur)
        p.next()
    }
    // Step through body parser to observe progress
    bp := &bodyParser{toks: body, comments: nil}
    for i := 0; i < 100; i++ { // prevent runaway
        if bp.atEnd() {
            break
        }
        cur := bp.cur()
        t.Logf("iter=%d i=%d cur.kind=%v lex=%q", i, bp.i, cur.Kind, cur.Lexeme)
        if _, ok := bp.parseStmt(); ok {
            t.Logf("stmt parsed at i=%d", bp.i)
            continue
        }
        // if not parsed, advance
        bp.next()
    }
}

func TestDebug_MinimalExpr(t *testing.T) {
    toks := []tok.Token{
        {Kind: tok.IDENT, Lexeme: "a"},
        {Kind: tok.PLUS, Lexeme: "+"},
        {Kind: tok.NUMBER, Lexeme: "2"},
    }
    bp := &bodyParser{toks: toks}
    if _, ok := bp.parseStmt(); !ok {
        t.Fatalf("parseStmt failed on minimal expr")
    }
}

func TestDebug_VarAssignExpr(t *testing.T) {
    toks := []tok.Token{
        {Kind: tok.KW_VAR, Lexeme: "var"},
        {Kind: tok.IDENT, Lexeme: "x"},
        {Kind: tok.ASSIGN, Lexeme: "="},
        {Kind: tok.IDENT, Lexeme: "a"},
        {Kind: tok.PLUS, Lexeme: "+"},
        {Kind: tok.NUMBER, Lexeme: "1"},
    }
    bp := &bodyParser{toks: toks}
    if _, ok := bp.parseStmt(); !ok {
        t.Fatalf("parseStmt failed on var assign expr")
    }
}

func TestDebug_MutAssignExpr(t *testing.T) {
    toks := []tok.Token{
        {Kind: tok.STAR, Lexeme: "*"},
        {Kind: tok.IDENT, Lexeme: "a"},
        {Kind: tok.ASSIGN, Lexeme: "="},
        {Kind: tok.IDENT, Lexeme: "a"},
        {Kind: tok.PLUS, Lexeme: "+"},
        {Kind: tok.NUMBER, Lexeme: "3"},
    }
    bp := &bodyParser{toks: toks}
    if _, ok := bp.parseStmt(); !ok {
        t.Fatalf("parseStmt failed on mut assign expr")
    }
}
