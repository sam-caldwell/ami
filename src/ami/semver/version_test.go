package semver

import "testing"

func TestParseVersion_AndCompare(t *testing.T) {
    a, err := ParseVersion("v1.2.3")
    if err != nil { t.Fatalf("parse: %v", err) }
    b, _ := ParseVersion("1.2.3")
    if Compare(a, b) != 0 { t.Fatalf("expected equal") }
    c, _ := ParseVersion("1.2.4")
    if Compare(c, b) <= 0 { t.Fatalf("expected c>b") }
    d, _ := ParseVersion("1.2.3-rc.1")
    if Compare(d, b) >= 0 { t.Fatalf("prerelease should be < release") }
}

