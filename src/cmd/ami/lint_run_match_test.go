package main

import "testing"

func Test_matchAnyRule(t *testing.T) {
    if !matchAnyRule("ABC_DEF", []string{"DEF"}) { t.Fatal("substring") }
    if !matchAnyRule("LINT_CODE", []string{"LINT_*"}) { t.Fatal("glob") }
    if !matchAnyRule("ERR123", []string{"re:^ERR\\d+$"}) { t.Fatal("regex") }
    if matchAnyRule("FOO", []string{"BAR"}) { t.Fatal("no match") }
}

