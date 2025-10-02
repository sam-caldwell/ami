package logger

import "io"

// StderrSink writes lines to stderr (or a provided writer).
type StderrSink struct { w io.Writer }

func (s *StderrSink) Start() error { return nil }

func (s *StderrSink) Write(line []byte) error {
    _, err := s.w.Write(line)
    return err
}

func (s *StderrSink) Close() error { return nil }

