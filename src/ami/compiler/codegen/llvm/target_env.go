package llvm

import "strings"

// TripleForEnv accepts an env string like "os/arch" and returns a triple.
func TripleForEnv(env string) string {
    parts := strings.SplitN(env, "/", 2)
    if len(parts) != 2 { return DefaultTriple }
    return TripleFor(parts[0], parts[1])
}

