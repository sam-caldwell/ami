package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// listGitTags returns tag names from a repo URL or local path using `git ls-remote --tags` (for remotes)
// or `git -C <path> tag` (for local file path). The returned tags are plain names (e.g., v1.2.3).
func listGitTags(cloneArg string) ([]string, error) {
    // Detect local path vs remote
    if filepath.IsAbs(cloneArg) {
        cmd := exec.Command("git", "-C", cloneArg, "tag")
        cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
        out, err := cmd.CombinedOutput()
        if err != nil { return nil, fmt.Errorf("git tag: %v: %s", err, string(out)) }
        lines := strings.Split(string(out), "\n")
        var tags []string
        for _, l := range lines { l = strings.TrimSpace(l); if l != "" { tags = append(tags, l) } }
        return tags, nil
    }
    cmd := exec.Command("git", "ls-remote", "--tags", cloneArg)
    cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_SSH_COMMAND=ssh -oBatchMode=yes -oStrictHostKeyChecking=no -oConnectTimeout=2")
    out, err := cmd.CombinedOutput()
    if err != nil { return nil, fmt.Errorf("git ls-remote: %v: %s", err, string(out)) }
    var tags []string
    for _, line := range strings.Split(string(out), "\n") {
        // lines like: <sha>\trefs/tags/v1.2.3
        if i := strings.Index(line, "refs/tags/"); i >= 0 {
            tag := line[i+len("refs/tags/"):]
            // strip ^{} suffix for annotated tags
            if j := strings.Index(tag, "^"); j > 0 { tag = tag[:j] }
            tag = strings.TrimSpace(tag)
            if tag != "" { tags = append(tags, tag) }
        }
    }
    return tags, nil
}

