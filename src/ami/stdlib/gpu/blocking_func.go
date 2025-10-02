package gpu

import "fmt"

// Blocking executes f and blocks until it returns. Any panic is converted to an error.
func Blocking(f func() error) (err error) {
    if f == nil { return fmt.Errorf("gpu: nil function") }
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("gpu: panic: %v", r)
        }
    }()
    return f()
}

