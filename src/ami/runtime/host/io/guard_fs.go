package io

import "errors"

// ErrCapabilityDenied is returned when an operation is blocked by I/O capability policy.
var ErrCapabilityDenied = errors.New("io: capability denied")

func guardFS() error { if !current.AllowFS { return ErrCapabilityDenied }; return nil }

