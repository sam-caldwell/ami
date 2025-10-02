package semver

import "testing"

func TestConstraint_Parse_Basic(t *testing.T) {
    if _, err := ParseConstraint("^1.2.3"); err != nil { t.Fatalf("parse: %v", err) }
}

