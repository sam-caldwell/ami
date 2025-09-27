package types

import "testing"

func TestKind_String_Values(t *testing.T) {
    cases := []struct{ k Kind; s string }{
        {Invalid, "invalid"},
        {Bool, "bool"},
        {Int, "int"},
        {Int64, "int64"},
        {Float64, "float64"},
        {String, "string"},
    }
    for _, c := range cases {
        if got := c.k.String(); got != c.s {
            t.Fatalf("%v.String()=%q want %q", c.k, got, c.s)
        }
    }
}

