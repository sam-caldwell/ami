package semver

import "testing"

func TestValidateVersion(t *testing.T) {
    if !ValidateVersion("1.2.3") { t.Fatalf("expected valid") }
}

