package workspace

import "testing"

func TestParseConstraint_AcceptsForms(t *testing.T) {
    cases := []string{"1.2.3", "v1.2.3", "^1.2.3", "~1.2.3", ">1.2.3", ">=1.2.3", ">= 1.2.3", "==latest"}
    for _, s := range cases {
        if _, err := ParseConstraint(s); err != nil {
            t.Fatalf("parse %q: %v", s, err)
        }
    }
}

func TestParseConstraint_RejectsUnsupported(t *testing.T) {
    bad := []string{"<=1.2.3", "!1.2.3", "foo", "^"}
    for _, s := range bad {
        if _, err := ParseConstraint(s); err == nil {
            t.Fatalf("expected error for %q", s)
        }
    }
}

func TestSatisfies_ExactAndRanges(t *testing.T) {
    // exact
    c, _ := ParseConstraint("v1.2.3")
    if !Satisfies("1.2.3", c) || Satisfies("1.2.4", c) { t.Fatalf("exact mismatch") }
    // > and >=
    gt, _ := ParseConstraint(">1.2.3")
    if Satisfies("1.2.3", gt) || !Satisfies("1.2.4", gt) { t.Fatalf("> failed") }
    gte, _ := ParseConstraint(">=1.2.3")
    if !Satisfies("1.2.3", gte) || !Satisfies("1.2.4", gte) { t.Fatalf(">= failed") }
    // ~1.2.3: >=1.2.3 <1.3.0
    tld, _ := ParseConstraint("~1.2.3")
    if !Satisfies("1.2.9", tld) || Satisfies("1.3.0", tld) { t.Fatalf("~ failed") }
    // ^1.2.3: >=1.2.3 <2.0.0
    crt, _ := ParseConstraint("^1.2.3")
    if !Satisfies("1.9.9", crt) || Satisfies("2.0.0", crt) { t.Fatalf("^ failed") }
    // ^0.2.3: >=0.2.3 <0.3.0
    crt2, _ := ParseConstraint("^0.2.3")
    if !Satisfies("0.2.9", crt2) || Satisfies("0.3.0", crt2) { t.Fatalf("^0.x failed") }
    // ^0.0.3: >=0.0.3 <0.0.4
    crt3, _ := ParseConstraint("^0.0.3")
    if !Satisfies("0.0.3", crt3) || Satisfies("0.0.4", crt3) { t.Fatalf("^0.0.x failed") }
    // latest
    lat, _ := ParseConstraint("==latest")
    if !Satisfies("99.0.0", lat) { t.Fatalf("latest should satisfy any version") }
}

