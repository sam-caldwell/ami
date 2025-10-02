package semver

import "testing"

func TestBound_Contains_FilePair(t *testing.T) {
    b := Bound{Lower: Version{1,2,3,""}, LowerInclusive: true, Upper: &Version{1,3,0,""}}
    if !Contains(b, Version{1,2,3,""}) { t.Fatalf("expected contains lower bound") }
}

