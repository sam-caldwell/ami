package main

import "testing"

func TestLinkExtraFlags_Darwin_DefaultAndPIE(t *testing.T) {
    // default darwin adds dead_strip
    f := linkExtraFlags("darwin/arm64", nil)
    if !contains(f, "-Wl,-dead_strip") {
        t.Fatalf("expected -Wl,-dead_strip for darwin; got %v", f)
    }
    // PIE on darwin uses -Wl,-pie
    f = linkExtraFlags("darwin/arm64", []string{"PIE"})
    if !contains(f, "-Wl,-pie") {
        t.Fatalf("expected -Wl,-pie for darwin PIE; got %v", f)
    }
}

func TestLinkExtraFlags_Linux_StaticPIEDCE(t *testing.T) {
    f := linkExtraFlags("linux/amd64", []string{"PIE", "static", "dce"})
    if !contains(f, "-pie") { t.Fatalf("expected -pie for linux PIE; got %v", f) }
    if !contains(f, "-static") { t.Fatalf("expected -static for linux; got %v", f) }
    if !contains(f, "-Wl,--gc-sections") { t.Fatalf("expected --gc-sections for linux dce; got %v", f) }
}

func contains(list []string, s string) bool {
    for _, x := range list { if x == s { return true } }
    return false
}

