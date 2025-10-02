package logger

import "io"

// StdoutSink writes lines to stdout (or a provided writer).
type StdoutSink struct { w io.Writer }

func (s *StdoutSink) Start() error { return nil }

func (s *StdoutSink) Write(line []byte) error {
    _, err := s.w.Write(line)
    return err
}

func (s *StdoutSink) Close() error { return nil }

