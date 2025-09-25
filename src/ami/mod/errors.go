package mod

import "errors"

// ErrNetwork is a sentinel error indicating a network/registry failure.
var ErrNetwork = errors.New("network registry error")
