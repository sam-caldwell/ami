package mod

import (
    "errors"
    "fmt"
    "net/url"
    "path/filepath"
    "strings"
    "os"

    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    gogitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type gitSSHBackend struct{}

func (gitSSHBackend) Name() string { return "git+ssh" }

func (gitSSHBackend) Match(spec string) bool { return strings.HasPrefix(spec, "git+ssh://") }

func (gitSSHBackend) Fetch(spec, cacheDir string) (string, string, string, error) {
    parts, err := parseGitPlusSSH(spec)
    if err != nil { return "", "", "", err }
    dest := filepath.Join(cacheDir, fmt.Sprintf("%s@%s", parts.Repo, parts.Tag))
    _ = removeAllNoFail(dest)
    auth, _ := gogitssh.NewSSHAgentAuth("git")
    repo, err := git.PlainClone(dest, false, &git.CloneOptions{URL: parts.SSHURL, Auth: auth, Depth: 1})
    if err != nil { return "", "", "", errors.Join(ErrNetwork, err) }
    wt, err := repo.Worktree()
    if err != nil { return "", "", "", err }
    if err := wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewTagReferenceName(parts.Tag)}); err != nil {
        return "", "", "", errors.Join(ErrNetwork, err)
    }
    // Compute package name host/path (no .git)
    raw := strings.TrimPrefix(spec, "git+")
    if i := strings.Index(raw, "#"); i >= 0 { raw = raw[:i] }
    pkg := ""
    if u, err := url.Parse(raw); err == nil {
        host := u.Host
        repoPath := strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), ".git")
        pkg = filepath.Join(host, repoPath)
    }
    return dest, pkg, parts.Tag, nil
}

// removeAllNoFail mirrors os.RemoveAll but ignores errors to simplify cleanup.
func removeAllNoFail(path string) error { _ = osRemoveAll(path); return nil }

// osRemoveAll is split for testability.
var osRemoveAll = func(path string) error { return osRemoveAllImpl(path) }

func osRemoveAllImpl(path string) error { return os.RemoveAll(path) }

func init() { registerBackend(gitSSHBackend{}) }
