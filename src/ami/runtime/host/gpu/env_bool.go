package gpu

import (
    "errors"
    "os"
    "strings"
)

// Sentinel errors used by the stubbed GPU stdlib.
var (
    ErrUnavailable     = errors.New("gpu: backend unavailable")
    ErrUnimplemented   = errors.New("gpu: unimplemented stub")
    ErrInvalidHandle   = errors.New("gpu: invalid handle")
    ErrAlreadyReleased = errors.New("gpu: already released")
)

// envBoolTrue returns true for common truthy strings (1,true,yes,on).
func envBoolTrue(name string) bool {
    v := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
    return v == "1" || v == "true" || v == "yes" || v == "on"
}

