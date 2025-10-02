package errors

import "testing"

func TestValidate_MinimalAndMissing(t *testing.T) {
    // minimal valid
    if err := Validate(Error{Level: "error", Code: "E", Message: "m"}); err != nil {
        t.Fatalf("expected valid error, got %v", err)
    }
    // missing fields
    if err := Validate(Error{}); err == nil {
        t.Fatalf("expected validation error for empty Error")
    }
}

