package mod

import "testing"

func TestSemverCompareAndMatch(t *testing.T) {
    if !isSemVer("v1.2.3") { t.Fatal("semver detect failed") }
    if semverLess("v1.2.3","v1.2.3") { t.Fatal("equal compare wrong") }
    if !semverLess("v1.2.3","v1.2.4") { t.Fatal("less compare wrong") }
    if semverLess("v2.0.0","v1.9.9") { t.Fatal("major compare wrong") }
}
