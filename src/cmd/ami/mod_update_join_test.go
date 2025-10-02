package main

import "testing"

func Test_joinCSV(t *testing.T) {
    if joinCSV(nil) != "" { t.Fatal("nil") }
}

