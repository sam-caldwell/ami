package testutil

import (
    "testing"
    "time"
)

func TestScaleInt_Simple(t *testing.T) {
    t.Setenv("AMI_TEST_TIMEOUT_SCALE", "2")
    if got := ScaleInt(3); got != 6 { t.Fatalf("ScaleInt(3) got %d want 6", got) }
}

func TestTimeout_Simple(t *testing.T) {
    t.Setenv("AMI_TEST_TIMEOUT_SCALE", "0.5")
    base := 2 * time.Second
    got := Timeout(base)
    if got != 1*time.Second { t.Fatalf("Timeout scale got %v want %v", got, 1*time.Second) }
}
