package workspace

import "testing"

func TestCmpSemver_Basic(t *testing.T) {
    a := semver{Major:1, Minor:0, Patch:0}
    b := semver{Major:1, Minor:0, Patch:1}
    if cmpSemver(a,b) >= 0 { t.Fatalf("expected a<b") }
}

