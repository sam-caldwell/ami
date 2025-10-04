package logger

import (
    "io"
    "os"
)

func NewStderrSink(w io.Writer) *StderrSink {
    if w == nil { w = os.Stderr }
    return &StderrSink{w: w}
}

