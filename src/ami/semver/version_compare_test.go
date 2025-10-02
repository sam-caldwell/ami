package semver

import "testing"

func TestVersion_Compare_Equal(t *testing.T) {
    a := Version{Major:1, Minor:2, Patch:3}
    b := Version{Major:1, Minor:2, Patch:3}
    if Compare(a,b) != 0 { t.Fatalf("expected equal") }
}

