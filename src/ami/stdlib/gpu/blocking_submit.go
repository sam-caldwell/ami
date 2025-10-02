package gpu

import "fmt"

// BlockingSubmit creates a completion channel, calls submit with it, then blocks
// waiting for an error result. Any panic in submit is converted to an error.
// The submit function must send exactly one error (which may be nil).
func BlockingSubmit(submit func(done chan<- error)) (err error) {
    if submit == nil { return fmt.Errorf("gpu: nil submit") }
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("gpu: panic: %v", r)
        }
    }()
    done := make(chan error, 1)
    submit(done)
    return <-done
}

