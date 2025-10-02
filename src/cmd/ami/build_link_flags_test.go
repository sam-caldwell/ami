package main

import "testing"

func TestLinkExtraFlags_FilePair(t *testing.T) {
    _ = linkExtraFlags("darwin/arm64", []string{"pie", "dead_strip"})
}

