package main

import "testing"

func TestIOAllowedIngress_Branches(t *testing.T) {
    trues := []string{
        "io.Stdin", // stdin
        "io.net.Listen", "io.net.Bind", "io.net.Accept", // listen/bind/accept
        "io.file.Read", "io.file.Open", // read/open without write
        "io.dir.ls", "io.dir.listdir", "io.dir.readdir", "io.dir.dirlist", // directory list
        "io.file.stat", // stat
        "io.file.seek", // seek
    }
    falses := []string{
        "io.file.write", "io.file.readwrite", // write-ish
        "io.unknown.op",
    }
    for _, s := range trues {
        if !ioAllowedIngress(s) {
            t.Fatalf("expected true for %q", s)
        }
    }
    for _, s := range falses {
        if ioAllowedIngress(s) {
            t.Fatalf("expected false for %q", s)
        }
    }
}
