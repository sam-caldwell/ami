package os

// prestartWriter buffers writes before Start(); Start will feed the buffer
// to the child process and then switch stdin to a live pipe.
type prestartWriter struct { p *Process }

func (w *prestartWriter) Write(b []byte) (int, error) {
    if w == nil || w.p == nil { return 0, errInvalidProcess }
    w.p.mu.Lock(); defer w.p.mu.Unlock()
    if w.p.started {
        // If Start already happened, forward to live pipe writer if available
        if w.p.stdin != nil { return w.p.stdin.Write(b) }
        return 0, errInvalidProcess
    }
    return w.p.stdinBuf.Write(b)
}

func (w *prestartWriter) Close() error {
    if w == nil || w.p == nil { return errInvalidProcess }
    w.p.mu.Lock(); w.p.preClosed = true; w.p.mu.Unlock()
    return nil
}

