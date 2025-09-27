package logger

import (
    "io"
    "os"
)

// StdoutSink writes lines to stdout (or a provided writer).
type StdoutSink struct { w io.Writer }

func NewStdoutSink(w io.Writer) *StdoutSink {
    if w == nil { w = os.Stdout }
    return &StdoutSink{w: w}
}

func (s *StdoutSink) Start() error { return nil }

func (s *StdoutSink) Write(line []byte) error {
    _, err := s.w.Write(line)
    return err
}

func (s *StdoutSink) Close() error { return nil }

