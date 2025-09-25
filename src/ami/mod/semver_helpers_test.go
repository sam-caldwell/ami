package mod

import "testing"

func TestPrereleaseAndParseSemVer(t *testing.T) {
    if !isSemVer("v1.2.3-rc.1") { t.Fatal("expected valid prerelease semver") }
    if prerelease("v1.2.3") != "" { t.Fatal("expected empty prerelease for release") }
    if prerelease("v1.2.3-rc.1") != "rc.1" { t.Fatalf("unexpected prerelease: %q", prerelease("v1.2.3-rc.1")) }
    a := parseSemVer("v1.2.3")
    if a != [3]int{1,2,3} { t.Fatalf("parseSemVer: %+v", a) }
}

func TestSortSemver(t *testing.T) {
    tags := []string{"v1.0.0", "v1.0.0-rc.1", "v1.0.1", "v2.0.0"}
    sortSemver(tags)
    want := []string{"v1.0.0-rc.1", "v1.0.0", "v1.0.1", "v2.0.0"}
    for i := range want {
        if tags[i] != want[i] { t.Fatalf("order[%d]: got %q want %q (all=%v)", i, tags[i], want[i], tags) }
    }
}

