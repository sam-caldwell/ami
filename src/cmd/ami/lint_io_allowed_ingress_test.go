package main

import "testing"

func TestIOAllowedIngress_FilePair(t *testing.T) {
    _ = ioAllowedIngress("io.stdin")
}

