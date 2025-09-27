package workspace

import "testing"

func TestHighestSatisfying_ChoosesMax_NoPrerelease(t *testing.T) {
    vers := []string{"v1.2.3", "1.3.0-rc.1", "v1.4.0"}
    c, _ := ParseConstraint("^1.0.0")
    got, ok := HighestSatisfying(vers, c, false)
    if !ok || got != "v1.4.0" {
        t.Fatalf("want v1.4.0, got %q ok=%v", got, ok)
    }
}

func TestHighestSatisfying_IncludesPrerelease_WhenAllowed(t *testing.T) {
    vers := []string{"1.4.0-rc.2", "1.4.0-rc.3"}
    c, _ := ParseConstraint("^1.0.0")
    got, ok := HighestSatisfying(vers, c, true)
    if !ok || got != "1.4.0-rc.3" {
        t.Fatalf("want rc.3, got %q ok=%v", got, ok)
    }
}

func TestHighestSatisfying_NoneFound(t *testing.T) {
    vers := []string{"0.1.0", "0.2.0"}
    c, _ := ParseConstraint(">1.0.0")
    if _, ok := HighestSatisfying(vers, c, false); ok {
        t.Fatalf("expected no match")
    }
}

