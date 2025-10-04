package io

import "os"

// CreateTempDir creates a unique temporary directory under the system temp dir and returns its path.
func CreateTempDir() (string, error) { if err := guardFS(); err != nil { return "", err }; return os.MkdirTemp(os.TempDir(), "ami-*") }

