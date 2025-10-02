package main

import "testing"

func TestIOAllowedEgress_FilePair(t *testing.T) {
    _ = ioAllowedEgress("io.stdout")
}

