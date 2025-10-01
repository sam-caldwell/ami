package os

import (
    goos "os"
    "strings"
)

// GetEnv returns the value of an environment variable (empty string if unset).
func GetEnv(name string) string {
    v, _ := goos.LookupEnv(name)
    return v
}

// SetEnv sets the value of an environment variable.
func SetEnv(name, value string) error { return goos.Setenv(name, value) }

// ListEnv returns a list of environment variable names (not their values).
func ListEnv() []string {
    env := goos.Environ()
    names := make([]string, 0, len(env))
    for _, kv := range env {
        if i := strings.IndexByte(kv, '='); i >= 0 { names = append(names, kv[:i]) }
    }
    return names
}
