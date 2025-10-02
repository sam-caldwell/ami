package main

import "testing"

func Test_anyKvEmit_falseWhenNoCases(t *testing.T) {
    if anyKvEmit(nil) { t.Fatal("expected false") }
}

