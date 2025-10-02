package os

import goos "os"

// SetEnv sets the value of an environment variable.
func SetEnv(name, value string) error { return goos.Setenv(name, value) }

