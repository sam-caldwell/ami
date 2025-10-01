package parser

import (
    "testing"
)

func TestSyntaxError_Error(t *testing.T) {
    // non-empty message
    if got := (SyntaxError{Msg: "boom"}).Error(); got != "boom" {
        t.Fatalf("Error(): want %q, got %q", "boom", got)
    }
    // empty message
    if got := (SyntaxError{}).Error(); got != "" {
        t.Fatalf("Error(): want empty string, got %q", got)
    }
}

