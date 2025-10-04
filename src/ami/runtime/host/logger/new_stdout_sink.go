package logger

import (
    "io"
    "os"
)

func NewStdoutSink(w io.Writer) *StdoutSink {
    if w == nil { w = os.Stdout }
    return &StdoutSink{w: w}
}

