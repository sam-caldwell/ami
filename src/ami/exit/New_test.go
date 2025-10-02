package exit

import (
    "errors"
    "testing"
)

// TestNew_ConstructsErrorWithCodeAndMessage verifies New returns an Error
// with the expected code and formatted message.
func TestNew_ConstructsErrorWithCodeAndMessage(t *testing.T) {
    err := New(User, "bad input: %d", 42)
    var e Error
    if !errors.As(err, &e) {
        t.Fatalf("expected exit.Error, got %T", err)
    }
    if e.Code != User {
        t.Fatalf("Code=%v, want %v", e.Code, User)
    }
    if e.Msg != "bad input: 42" {
        t.Fatalf("Msg=%q, want %q", e.Msg, "bad input: 42")
    }

    // Also validate without formatting args
    err2 := New(OK, "all good")
    var e2 Error
    if !errors.As(err2, &e2) || e2.Code != OK || e2.Msg != "all good" {
        t.Fatalf("unexpected result: %#v", e2)
    }
}
