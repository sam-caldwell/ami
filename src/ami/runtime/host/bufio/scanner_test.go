package bufio

import (
    "testing"
)

func TestScanner_Lines_Text_Bytes_Err(t *testing.T) {
    r, err := NewReader("a\nb\nc\n")
    if err != nil { t.Fatalf("NewReader: %v", err) }
    sc, err := NewScanner(r)
    if err != nil { t.Fatalf("NewScanner: %v", err) }

    var lines []string
    for sc.Scan() {
        lines = append(lines, sc.Text())
        b := sc.Bytes()
        bs, _ := b.Bytes()
        if len(bs) != len(sc.Text()) { t.Fatalf("bytes/text len mismatch: %d vs %d", len(bs), len(sc.Text())) }
        _ = b.Release()
    }
    if err := sc.Err(); err != nil { t.Fatalf("scanner err: %v", err) }
    if len(lines) != 3 || lines[0] != "a" || lines[1] != "b" || lines[2] != "c" {
        t.Fatalf("lines: %+v", lines)
    }
}

