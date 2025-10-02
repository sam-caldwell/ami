package main

import "testing"

func Test_modGetResult_zero(t *testing.T) {
    var r modGetResult
    if r.Name != "" || r.Version != "" || r.Source != "" || r.Path != "" || r.Message != "" {
        t.Fatalf("unexpected non-zero fields: %+v", r)
    }
}

