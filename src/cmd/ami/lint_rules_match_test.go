package main

import "testing"

func TestMatchAnyRule_Substring(t *testing.T) {
    if !matchAnyRule("W_FOO_BAR", []string{"FOO"}) { t.Fatalf("expected substring match") }
    if matchAnyRule("W_FOO_BAR", []string{"BAZ"}) { t.Fatalf("unexpected substring match") }
}

func TestMatchAnyRule_Glob(t *testing.T) {
    if !matchAnyRule("W_IDENT_UNDERSCORE", []string{"W_IDENT_*"}) { t.Fatalf("expected glob match") }
    if matchAnyRule("W_IDENT_UNDERSCORE", []string{"W_*_X"}) { t.Fatalf("unexpected glob match") }
}

func TestMatchAnyRule_RegexPrefix(t *testing.T) {
    if !matchAnyRule("E_PARSE", []string{"re:^E_.*$"}) { t.Fatalf("expected regex prefix match") }
    if matchAnyRule("E_PARSE", []string{"re:^W_.*$"}) { t.Fatalf("unexpected regex prefix match") }
}

func TestMatchAnyRule_RegexSlashes(t *testing.T) {
    if !matchAnyRule("W_TODO", []string{"/^W_.*/"}) { t.Fatalf("expected slash-regex match") }
    if matchAnyRule("W_TODO", []string{"/^E_.*/"}) { t.Fatalf("unexpected slash-regex match") }
}

