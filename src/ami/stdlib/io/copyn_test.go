package io

import (
	stdbytes "bytes"
	"testing"
)

func TestIO_CopyN_Happy(t *testing.T) {
	src := stdbytes.NewBufferString("hello world")
	var dst stdbytes.Buffer
	n, err := CopyN(&dst, src, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 || dst.String() != "hello" {
		t.Fatalf("copied %d %q", n, dst.String())
	}
}

func TestIO_CopyN_ShortSrc_Sad(t *testing.T) {
	src := stdbytes.NewBufferString("hi")
	var dst stdbytes.Buffer
	n, err := CopyN(&dst, src, 5)
	if err == nil {
		t.Fatal("expected error due to short read")
	}
	if n != 2 || dst.String() != "hi" {
		t.Fatalf("copied %d %q", n, dst.String())
	}
}
