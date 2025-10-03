package main

import "testing"

func TestLinkExtraFlags_Mapping(t *testing.T) {
    got := linkExtraFlags("darwin/arm64", []string{"pie", "dead_strip"})
    has := func(xs []string, s string) bool { for _, x := range xs { if x == s { return true } }; return false }
    if !has(got, "-Wl,-dead_strip") || !has(got, "-Wl,-pie") || !has(got, "-framework") {
        t.Fatalf("darwin flags missing expected entries: %v", got)
    }
    got2 := linkExtraFlags("linux/amd64", []string{"static", "dead_strip"})
    if !has(got2, "-static") || !has(got2, "-Wl,--gc-sections") {
        t.Fatalf("linux flags missing expected entries: %v", got2)
    }
}
