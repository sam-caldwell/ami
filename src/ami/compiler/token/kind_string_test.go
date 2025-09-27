package token

import "testing"

func TestKindString(t *testing.T) {
    cases := map[Kind]string{
        Unknown:   "Unknown",
        EOF:       "EOF",
        Ident:     "Ident",
        Number:    "Number",
        String:    "String",
        Symbol:    "Symbol",
        Assign:    "Assign",
        Eq:        "Eq",
        KwPackage: "KwPackage",
    }
    for k, want := range cases {
        if got := k.String(); got != want {
            t.Fatalf("Kind(%d).String() => %q; want %q", k, got, want)
        }
    }
}
