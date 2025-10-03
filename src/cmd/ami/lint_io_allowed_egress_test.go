package main

import "testing"

func TestIOAllowedEgress_Branches(t *testing.T) {
    trues := []string{
        "io.Stdout", "io.Stderr", // stdio
        "io.net.Connect", "io.net.Dial", // connect/dial
        "io.net.udp.send", "io.net.tcp.send", "io.net.icmp.sendto", // send variants (with protocol hints)
        "io.file.Write", "io.file.Append", // write/append
        "io.file.Create", "io.file.Delete", "io.file.Truncate", "io.file.Chmod", "io.file.Chown", // file mutate
        "io.file.Stat", "io.file.Read", "io.file.Seek", // file stat/read/seek
        "io.fs.Mkdir", "io.fs.MkdirAll", "io.fs.DirCreate", "io.fs.Rmdir", "io.fs.DirDelete", // dir ops
        "io.tmp.TempFile", "io.tmp.CreateTemp", "io.tmp.TempDir", "io.tmp.CreateTempDir", // temp ops
    }
    falses := []string{
        "io.unknown", // unknown op
    }
    for _, s := range trues {
        if !ioAllowedEgress(s) { t.Fatalf("expected true for %q", s) }
    }
    for _, s := range falses {
        if ioAllowedEgress(s) { t.Fatalf("expected false for %q", s) }
    }
}
