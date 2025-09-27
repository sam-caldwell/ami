package semver

import "testing"

func TestBounds_CaretAndTilde(t *testing.T) {
    c, _ := ParseConstraint("^1.2.3")
    b, ok := Bounds(c)
    if !ok || !b.LowerInclusive || b.Upper == nil || b.Upper.Major != 2 || b.UpperInclusive {
        t.Fatalf("caret bounds: %+v ok=%v", b, ok)
    }
    c2, _ := ParseConstraint("~1.2.3")
    b2, ok := Bounds(c2)
    if !ok || b2.Upper == nil || b2.Upper.Major != 1 || b2.Upper.Minor != 3 || b2.UpperInclusive {
        t.Fatalf("tilde bounds: %+v ok=%v", b2, ok)
    }
}

func TestIntersect_And_Contains(t *testing.T) {
    vA, _ := ParseVersion("1.2.3")
    vB, _ := ParseVersion("1.3.0")
    b1 := Bound{Lower: vA, LowerInclusive: true, Upper: &vB, UpperInclusive: false}
    vC, _ := ParseVersion("1.2.4")
    vD, _ := ParseVersion("1.2.5")
    b2 := Bound{Lower: vC, LowerInclusive: true, Upper: &vD, UpperInclusive: true}
    bi, ok := Intersect(b1, b2)
    if !ok { t.Fatalf("expected non-empty intersection") }
    if !Contains(bi, vC) || Contains(bi, vB) { t.Fatalf("contains checks failed: %+v", bi) }
}

