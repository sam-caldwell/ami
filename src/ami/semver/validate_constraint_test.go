package semver

import "testing"

func TestValidateConstraint(t *testing.T) {
    if !ValidateConstraint("1.2.3") { t.Fatalf("expected valid") }
}

