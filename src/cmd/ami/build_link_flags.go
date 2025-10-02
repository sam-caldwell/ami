package main

import "strings"

// linkExtraFlags returns a set of linker flags adjusted for the target env
// and workspace linker options.
func linkExtraFlags(env string, opts []string) []string {
    var extra []string
    // Default: on Darwin, prefer dead strip
    if env == "darwin/arm64" || env == "darwin/amd64" || env == "darwin/x86_64" {
        extra = append(extra, "-Wl,-dead_strip")
        // Link against Apple frameworks needed for Metal runtime integrations
        extra = append(extra, "-framework", "Foundation", "-framework", "Metal")
    }
    // Options mapping
    for _, opt := range opts {
        switch opt {
        case "PIE", "pie":
            if env == "darwin/arm64" || env == "darwin/amd64" || env == "darwin/x86_64" {
                extra = append(extra, "-Wl,-pie")
            } else {
                extra = append(extra, "-pie")
            }
        case "static":
            // Best effort: static commonly supported on Linux
            if strings.HasPrefix(env, "linux/") {
                extra = append(extra, "-static")
            }
        case "dead_strip", "dce":
            if strings.HasPrefix(env, "darwin/") {
                extra = append(extra, "-Wl,-dead_strip")
            }
            if strings.HasPrefix(env, "linux/") {
                extra = append(extra, "-Wl,--gc-sections")
            }
        }
    }
    return extra
}

