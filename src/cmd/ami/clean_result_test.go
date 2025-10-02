package main

import "testing"

func Test_cleanResult_zero(t *testing.T) {
    var r cleanResult
    if r.Path != "" || r.Removed || r.Created || r.Messages != nil { t.Fatalf("unexpected: %+v", r) }
}

