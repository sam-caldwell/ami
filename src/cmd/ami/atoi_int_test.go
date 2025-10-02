package main

import "testing"

func Test_atoi(t *testing.T) {
    if v, _ := atoi("42"); v != 42 { t.Fatalf("got %d", v) }
}

