package token

import "testing"

func TestOperators_Lookup(t *testing.T) {
    if Operators["+"] != Plus {
        t.Fatalf("Operators[+] => %v; want Plus", Operators["+"])
    }
    if Operators["=="] != Eq {
        t.Fatalf("Operators[==] => %v; want Eq", Operators["=="])
    }
    if _, ok := Operators["??"]; ok {
        t.Fatalf("unexpected operator mapping for ??")
    }
}

