package os

import (
    goos "os"
    "strings"
)

// ListEnv returns a list of environment variable names (not their values).
func ListEnv() []string {
    env := goos.Environ()
    names := make([]string, 0, len(env))
    for _, kv := range env {
        if i := strings.IndexByte(kv, '='); i >= 0 { names = append(names, kv[:i]) }
    }
    return names
}

