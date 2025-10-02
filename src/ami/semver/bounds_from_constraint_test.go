package semver

import "testing"

func TestBounds_FromConstraint_FilePair(t *testing.T) {
    c, _ := ParseConstraint("^1.2.3")
    if _, ok := Bounds(c); !ok { t.Fatalf("expected ok bounds") }
}

