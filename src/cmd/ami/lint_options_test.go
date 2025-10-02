package main

import "testing"

func Test_LintOptions_defaultsReset(t *testing.T) {
    setLintOptions(LintOptions{Rules: []string{"X"}, MaxWarn: 5, FailFast: true})
    if currentLintOptions.MaxWarn != 5 || !currentLintOptions.FailFast { t.Fatal("precondition failed") }
    // reset
    setLintOptions(LintOptions{})
    if currentLintOptions.MaxWarn != -1 || currentLintOptions.FailFast { t.Fatalf("defaults not restored: %+v", currentLintOptions) }
}

