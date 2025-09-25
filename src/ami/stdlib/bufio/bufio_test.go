package bufio

import (
    stdbytes "bytes"
    stdio "io"
    "testing"
)

type errWriter struct{}
func (errWriter) Write(p []byte) (int, error) { return 0, stdio.ErrShortWrite }

func TestBufio_Writer_FlushSemantics(t *testing.T) {
    var underlying stdbytes.Buffer
    w, err := NewWriterSize(&underlying, 16)
    if err != nil { t.Fatal(err) }
    if _, err := w.Write([]byte("hello")); err != nil { t.Fatal(err) }
    if underlying.Len() != 0 { t.Fatalf("unexpected write-through before flush: %d", underlying.Len()) }
    if err := w.Flush(); err != nil { t.Fatal(err) }
    if underlying.String() != "hello" { t.Fatalf("flush got %q", underlying.String()) }
}

func TestBufio_Writer_BufferFull_PartialFlush(t *testing.T) {
    var underlying stdbytes.Buffer
    w, err := NewWriterSize(&underlying, 3)
    if err != nil { t.Fatal(err) }
    if _, err := w.Write([]byte("ab")); err != nil { t.Fatal(err) }
    if underlying.Len() != 0 { t.Fatalf("unexpected write-through before full: %q", underlying.String()) }
    if _, err := w.Write([]byte("c")); err != nil { t.Fatal(err) }
    if underlying.Len() != 0 { t.Fatalf("unexpected flush at exactly full: %q", underlying.String()) }
    if _, err := w.Write([]byte("d")); err != nil { t.Fatal(err) }
    if underlying.String() != "abc" { t.Fatalf("expected flush when exceeding buffer, got %q", underlying.String()) }
    if _, err := w.Write([]byte("e")); err != nil { t.Fatal(err) }
    if err := w.Flush(); err != nil { t.Fatal(err) }
    if underlying.String() != "abcde" { t.Fatalf("after flush got %q", underlying.String()) }
}

func TestBufio_Writer_FlushError(t *testing.T) {
    w, err := NewWriterSize(errWriter{}, 4)
    if err != nil { t.Fatal(err) }
    if _, err := w.Write([]byte("abc")); err != nil { t.Fatal(err) }
    if err := w.Flush(); err == nil { t.Fatal("expected flush error") }
}

func TestBufio_Reader_Read(t *testing.T) {
    src := stdbytes.NewBufferString("abcdef")
    r, err := NewReaderSize(src, 2)
    if err != nil { t.Fatal(err) }
    buf := make([]byte, 4)
    n, err := r.Read(buf)
    if err != nil && err != stdio.EOF { t.Fatal(err) }
    if n <= 0 { t.Fatal("expected some bytes") }
}

func TestBufio_InvalidSizes(t *testing.T) {
    if _, err := NewReaderSize(stdbytes.NewBuffer(nil), 0); err == nil { t.Fatal("expected error") }
    var b stdbytes.Buffer
    if _, err := NewWriterSize(&b, -1); err == nil { t.Fatal("expected error") }
}
