package main

import "testing"

func TestSelectHighestSemver_IncludingPrerelease(t *testing.T) {
    tags := []string{"v1.2.3", "v1.10.0", "v2.0.0-rc.1", "bad", "v0.9.9"}
    got, err := selectHighestSemver(tags, true)
    if err != nil { t.Fatalf("selectHighestSemver: %v", err) }
    if got != "v2.0.0-rc.1" { t.Fatalf("want v2.0.0-rc.1, got %s", got) }
}

func TestSelectHighestSemver_ExcludePrerelease(t *testing.T) {
    tags := []string{"v1.2.3", "v1.10.0", "v2.0.0-rc.1"}
    got, err := selectHighestSemver(tags, false)
    if err != nil { t.Fatalf("selectHighestSemver: %v", err) }
    if got != "v1.10.0" { t.Fatalf("want v1.10.0, got %s", got) }
}

func TestAtoi_WrapsStrconv(t *testing.T) {
    if _, err := atoi("123"); err != nil { t.Fatalf("atoi valid: %v", err) }
    if _, err := atoi("x"); err == nil { t.Fatalf("expected error for invalid atoi input") }
}

