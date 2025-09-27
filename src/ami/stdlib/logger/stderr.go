package logger

import (
    "io"
    "os"
)

// StderrSink writes lines to stderr (or a provided writer).
type StderrSink struct { w io.Writer }

func NewStderrSink(w io.Writer) *StderrSink {
    if w == nil { w = os.Stderr }
    return &StderrSink{w: w}
}

func (s *StderrSink) Start() error { return nil }

func (s *StderrSink) Write(line []byte) error {
    _, err := s.w.Write(line)
    return err
}

func (s *StderrSink) Close() error { return nil }

