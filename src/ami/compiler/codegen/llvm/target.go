package llvm

import "strings"

// TripleFor maps an os/arch pair to an LLVM target triple string.
// Unknown pairs return DefaultTriple for compatibility.
func TripleFor(osName, arch string) string {
    osName = strings.ToLower(osName)
    arch = strings.ToLower(arch)
    switch osName + "/" + arch {
    case "darwin/arm64":
        return "arm64-apple-macosx"
    case "darwin/amd64", "darwin/x86_64":
        return "x86_64-apple-macosx"
    case "linux/arm64", "linux/aarch64":
        return "aarch64-unknown-linux-gnu"
    case "linux/amd64", "linux/x86_64":
        return "x86_64-unknown-linux-gnu"
    default:
        return DefaultTriple
    }
}

// TripleForEnv accepts an env string like "os/arch" and returns a triple.
func TripleForEnv(env string) string {
    parts := strings.SplitN(env, "/", 2)
    if len(parts) != 2 { return DefaultTriple }
    return TripleFor(parts[0], parts[1])
}

