package semver

import "testing"

func TestBound_Intersect_FilePair(t *testing.T) {
    a := Bound{Lower: Version{1,2,3,""}, LowerInclusive: true, Upper: &Version{1,3,0,""}}
    b := Bound{Lower: Version{1,2,4,""}, LowerInclusive: true, Upper: &Version{1,2,5,""}, UpperInclusive: true}
    if _, ok := Intersect(a,b); !ok { t.Fatalf("expected non-empty") }
}

