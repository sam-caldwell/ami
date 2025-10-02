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
        case "windows/amd64", "windows/x86_64":
            return "x86_64-pc-windows-msvc"
        case "windows/arm64":
            return "aarch64-pc-windows-msvc"
        case "linux/arm64", "linux/aarch64":
            return "aarch64-unknown-linux-gnu"
        case "linux/amd64", "linux/x86_64":
            return "x86_64-unknown-linux-gnu"
        case "linux/riscv64":
            return "riscv64-unknown-linux-gnu"
        case "linux/arm":
            return "arm-unknown-linux-gnueabihf"
        case "freebsd/amd64", "freebsd/x86_64":
            return "x86_64-unknown-freebsd"
        case "freebsd/arm64", "freebsd/aarch64":
            return "aarch64-unknown-freebsd"
        case "openbsd/amd64", "openbsd/x86_64":
            return "x86_64-unknown-openbsd"
        case "openbsd/arm64", "openbsd/aarch64":
            return "aarch64-unknown-openbsd"
        default:
            return DefaultTriple
    }
}

// TripleForEnv accepts an env string like "os/arch" and returns a triple.
// TripleForEnv moved to target_env.go to satisfy single-declaration rule
