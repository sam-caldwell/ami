package io

import "os"

// Hostname returns the current system hostname.
func Hostname() (string, error) { if err := guardNet(); err != nil { return "", err }; return os.Hostname() }

