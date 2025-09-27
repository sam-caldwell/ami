package main

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "os/exec"
    "strconv"
    "strings"
)

// computeCommitDigest returns a deterministic SHA-256 digest for the commit
// identified by tagOrRef in the given git repository path. If the repository
// uses Git SHA-256 object format, the commit ID itself (64-hex) is returned.
// Otherwise, the digest of the canonical object content is computed as:
//   sha256("commit <len>\x00" + <body>)
func computeCommitDigest(repoPath, tagOrRef string) (string, error) {
    // Resolve commit id for the tag/ref
    cmd := exec.Command("git", "-C", repoPath, "rev-parse", tagOrRef+"^{commit}")
    cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")
    out, err := cmd.CombinedOutput()
    if err != nil { return "", fmt.Errorf("rev-parse: %v: %s", err, string(out)) }
    id := strings.TrimSpace(string(out))
    if len(id) == 64 {
        // Repository in SHA-256 object format: use commit ID directly
        return id, nil
    }
    // Obtain raw commit body
    cmd = exec.Command("git", "-C", repoPath, "cat-file", "-p", id)
    cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")
    body, err := cmd.CombinedOutput()
    if err != nil { return "", fmt.Errorf("cat-file: %v: %s", err, string(body)) }
    // Build canonical header
    header := []byte("commit " + strconv.Itoa(len(body)) + "\x00")
    h := sha256.New()
    _, _ = h.Write(header)
    _, _ = h.Write(body)
    return hex.EncodeToString(h.Sum(nil)), nil
}

