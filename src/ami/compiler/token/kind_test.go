package token

import "testing"

func TestKind_String(t *testing.T) {
    cases := []struct{ k Kind; s string }{
        {Unknown, "Unknown"},
        {Ident, "Ident"},
        {Number, "Number"},
        {String, "String"},
        {Symbol, "Symbol"},
        {EOF, "EOF"},
        {Kind(999), "Unknown"},
    }
    for _, c := range cases {
        if got := c.k.String(); got != c.s {
            t.Fatalf("%v.String()=%q want %q", int(c.k), got, c.s)
        }
    }
}

