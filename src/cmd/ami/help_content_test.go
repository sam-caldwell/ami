package main

import "testing"

func TestGetHelpDoc_ContainsHeader(t *testing.T) {
    s := getHelpDoc()
    if len(s) == 0 || s[:8] != "AMI Help" { t.Fatalf("help header missing") }
}

