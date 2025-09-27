package semver

import "testing"

func TestSatisfies(t *testing.T) {
    c, _ := ParseConstraint(">= v1.2.3")
    if !Satisfies("1.2.3", c) || !Satisfies("1.3.0", c) || Satisfies("1.2.2", c) { t.Fatalf(">=") }
    c2, _ := ParseConstraint("> v1.2.3")
    if !Satisfies("1.2.4", c2) || Satisfies("1.2.3", c2) { t.Fatalf(">") }
    c3, _ := ParseConstraint("^1.2.3")
    if !Satisfies("1.4.0", c3) || Satisfies("2.0.0", c3) || Satisfies("1.2.2", c3) { t.Fatalf("^") }
    c4, _ := ParseConstraint("~1.2.3")
    if !Satisfies("1.2.9", c4) || Satisfies("1.3.0", c4) || Satisfies("1.2.2", c4) { t.Fatalf("~") }
}

