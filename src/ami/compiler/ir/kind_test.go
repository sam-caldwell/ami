package ir

import "testing"

func TestKind_String(t *testing.T) {
    cases := []struct{ k Kind; s string }{
        {OpVar, "VAR"},
        {OpAssign, "ASSIGN"},
        {OpReturn, "RETURN"},
        {OpDefer, "DEFER"},
        {OpExpr, "EXPR"},
        {Kind(999), "?"},
    }
    for _, c := range cases {
        if got := c.k.String(); got != c.s {
            t.Fatalf("%v.String()=%q want %q", c.k, got, c.s)
        }
    }
}

