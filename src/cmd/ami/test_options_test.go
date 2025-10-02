package main

import "testing"

func Test_setTestOptions_sets(t *testing.T) {
    setTestOptions(TestOptions{TimeoutMs: 123})
    if currentTestOptions.TimeoutMs != 123 { t.Fatal("not set") }
}

