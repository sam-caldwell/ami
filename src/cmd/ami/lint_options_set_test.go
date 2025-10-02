package main

import "testing"

func Test_setLintOptions_setsFields(t *testing.T) {
    setLintOptions(LintOptions{}) // start from defaults
    setLintOptions(LintOptions{Rules: []string{"A","B"}, MaxWarn: 1, CompatCodes: true})
    if currentLintOptions.MaxWarn != 1 || !currentLintOptions.CompatCodes { t.Fatalf("unexpected: %+v", currentLintOptions) }
    if len(currentLintOptions.Rules) != 2 { t.Fatalf("rules: %v", currentLintOptions.Rules) }
}

