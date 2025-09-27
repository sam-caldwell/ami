package token

import "testing"

func TestSymbols_Basics(t *testing.T) {
    if LParen != "(" || RParen != ")" { t.Fatalf("paren constants incorrect") }
    if LBrace != "{" || RBrace != "}" { t.Fatalf("brace constants incorrect") }
    if Comma != "," || Semi != ";" { t.Fatalf("comma/semi constants incorrect") }
    if Dot != "." || Colon != ":" { t.Fatalf("dot/colon constants incorrect") }
    if Pipe != "|" || BackSlash != "\\" { t.Fatalf("pipe/backslash constants incorrect") }
    if Dollar != "$" || Exclamation != "!" { t.Fatalf("dollar/exclamation constants incorrect") }
}
