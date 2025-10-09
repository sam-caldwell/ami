package bufio

import (
    "bytes"
    "testing"
)

func TestReader_Peek_Read_UnreadByte_OwnedRelease(t *testing.T) {
    r, err := NewReader("hello world")
    if err != nil { t.Fatalf("NewReader: %v", err) }

    // Peek does not consume
    pk, err := r.Peek(5)
    if err != nil { t.Fatalf("Peek: %v", err) }
    b, _ := pk.Bytes()
    if string(b) != "hello" { t.Fatalf("peek got %q", string(b)) }
    if err := pk.Release(); err != nil { t.Fatalf("release: %v", err) }
    if err := pk.Release(); err != ErrAlreadyReleased { t.Fatalf("double release = %v", err) }

    // Read consumes
    h1, err := r.Read(6)
    if err != nil { t.Fatalf("Read(6): %v", err) }
    b1, _ := h1.Bytes()
    if string(b1) != "hello " { t.Fatalf("read1 got %q", string(b1)) }
    _ = h1.Release()

    // UnreadByte puts back last byte
    if err := r.UnreadByte(); err != nil { t.Fatalf("UnreadByte: %v", err) }
    h2, err := r.Read(6)
    if err != nil { t.Fatalf("Read(6) after unread: %v", err) }
    b2, _ := h2.Bytes()
    if string(b2) != " world" { t.Fatalf("read2 got %q", string(b2)) }
    _ = h2.Release()

    h3, err := r.Read(5)
    if err != nil { t.Fatalf("Read(5): %v", err) }
    b3, _ := h3.Bytes()
    if string(b3) != "" { t.Fatalf("read3 tail got %q", string(b3)) }
    _ = h3.Release()
}

func TestWriter_Write_Flush(t *testing.T) {
    var sink bytes.Buffer
    w, err := NewWriter(&sink)
    if err != nil { t.Fatalf("NewWriter: %v", err) }
    part1 := newOwnedBytes([]byte("abc"))
    n, err := w.Write(part1)
    if err != nil || n != 3 { t.Fatalf("Write part1: n=%d err=%v", n, err) }
    _ = part1.Release()

    part2 := newOwnedBytes([]byte("def"))
    n, err = w.Write(part2)
    if err != nil || n != 3 { t.Fatalf("Write part2: n=%d err=%v", n, err) }
    _ = part2.Release()

    // Before flush, sink should be empty
    if sink.Len() != 0 { t.Fatalf("sink len before flush = %d", sink.Len()) }
    if err := w.Flush(); err != nil { t.Fatalf("Flush: %v", err) }
    if got := sink.String(); got != "abcdef" { t.Fatalf("sink got %q", got) }
}
