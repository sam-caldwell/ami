package root

import "testing"

func TestSatisfiesConstraint(t *testing.T) {
	cases := []struct {
		ver, cons string
		ok        bool
	}{
		{"0.1.0", "==latest", true},
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "v1.2.3", true},
		{"1.2.4", ">1.2.3", true},
		{"1.2.3", ">1.2.3", false},
		{"1.3.0", ">=1.2.3", true},
		{"1.2.3", ">=1.2.3", true},
		{"2.0.0", "^1.2.3", false},
		{"1.3.0", "^1.2.3", true},
		{"1.2.5", "~1.2.3", true},
		{"1.3.0", "~1.2.3", false},
	}
	for _, c := range cases {
		if got := satisfiesConstraint(c.ver, c.cons); got != c.ok {
			t.Fatalf("ver=%s cons=%s got=%v want=%v", c.ver, c.cons, got, c.ok)
		}
	}
}
