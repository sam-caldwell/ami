package token

import "testing"

func TestPrecedence_Order(t *testing.T) {
    if !(Precedence(Or) < Precedence(And)) {
        t.Fatalf("expected Or < And precedence")
    }
    if !(Precedence(Eq) < Precedence(Lt)) {
        t.Fatalf("expected Eq < Rel precedence")
    }
    if !(Precedence(Plus) < Precedence(Star)) {
        t.Fatalf("expected Add < Mul precedence")
    }
    if Precedence(Ident) != 0 {
        t.Fatalf("expected non-operator to have precedence 0")
    }
}

