package main

import (
    "fmt"
    "os"
    "path/filepath"
    "os/exec"
)

// fetchGitToCache clones repo at tag into dest directory.
func fetchGitToCache(source, tag, dest string) error {
    // Parse source and convert file+git to absolute path
    repoURL := source
    cloneArg := repoURL
    if hasPrefix(repoURL, "file+git://") {
        cloneArg = repoURL[len("file+git://"):]
        if cloneArg == "" || !filepath.IsAbs(cloneArg) {
            return fmt.Errorf("file+git requires absolute path")
        }
    }
    tmp, err := os.MkdirTemp("", "ami-modsum-")
    if err != nil { return err }
    defer os.RemoveAll(tmp)
    env := os.Environ()
    env = append(env, "GIT_TERMINAL_PROMPT=0")
    env = append(env, "GIT_SSH_COMMAND=ssh -oBatchMode=yes -oStrictHostKeyChecking=no -oConnectTimeout=2")
    cmd := exec.Command("git", "clone", "--depth", "1", "--branch", tag, cloneArg, tmp)
    cmd.Env = env
    if err := cmd.Run(); err != nil { return err }
    // Copy into dest
    if err := os.RemoveAll(dest); err != nil { return err }
    return copyDir(tmp, dest)
}

