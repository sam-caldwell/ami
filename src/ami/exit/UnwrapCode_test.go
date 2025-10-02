package exit

import "testing"

func TestUnwrapCode(t *testing.T) {
    if got := UnwrapCode(nil); got != OK {
        t.Fatalf("expected OK, got %v", got)
    }
    err := New(User, "bad input")
    if got := UnwrapCode(err); got != User {
        t.Fatalf("expected User, got %v", got)
    }
    if OK.Int() != 0 || Internal.Int() != 1 {
        t.Fatalf("unexpected Int values: OK=%d Internal=%d", OK.Int(), Internal.Int())
    }
}

