package semver

import "testing"

func TestVersion_Parse_Basic(t *testing.T) {
    if _, err := ParseVersion("1.2.3"); err != nil { t.Fatalf("parse: %v", err) }
}

