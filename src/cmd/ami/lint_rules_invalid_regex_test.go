package main

import "testing"

func TestMatchAnyRule_InvalidRegex_NoMatch(t *testing.T) {
    if matchAnyRule("W_TODO", []string{"re:("}) { t.Fatalf("invalid regex should not match") }
}

