package os

import goos "os"

// GetEnv returns the value of an environment variable (empty string if unset).
func GetEnv(name string) string {
    v, _ := goos.LookupEnv(name)
    return v
}

