package main

import "testing"

func Test_modSumResult_zero(t *testing.T) {
    var r modSumResult
    if r.Path != "" || r.Ok || r.PackagesSeen != 0 || r.Schema != "" || r.Verified != nil || r.Missing != nil || r.Mismatched != nil || r.Message != "" {
        t.Fatalf("unexpected non-zero fields: %+v", r)
    }
}
