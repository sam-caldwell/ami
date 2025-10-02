package semver

import "testing"

func testParseConstraint_Valid(t *testing.T) {
	cases := []string{
		"v1.2.3",
		"1.2.3",
		"^1.2.3",
		"~1.2.3",
		">= 1.2.3",
		">1.2.3",
		"==latest",
		"1.2.3-rc.1",
	}
	for _, c := range cases {
		if _, err := ParseConstraint(c); err != nil {
			t.Fatalf("valid constraint rejected: %s: %v", c, err)
		}
	}
}

func testParseConstraint_Invalid(t *testing.T) {
	cases := []string{
		"",
		"<= 1.2.3",
		"1.2",
		"1",
		"abc",
		"v1.2.3.4",
	}
	for _, c := range cases {
		if _, err := ParseConstraint(c); err == nil {
			t.Fatalf("invalid constraint accepted: %s", c)
		}
	}
}
