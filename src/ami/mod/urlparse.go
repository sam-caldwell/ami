package mod

import (
    "errors"
    "net/url"
    "path/filepath"
    "strings"
)

// gitSSH holds parsed components of a git+ssh URL with semver tag.
type gitSSH struct {
    SSHURL  string // ssh://git@host/path.git
    Host    string // host
    Path    string // /org/repo.git
    Repo    string // repo base name (no .git)
    Tag     string // vX.Y.Z
}

// parseGitPlusSSH parses git+ssh://host/path#<tag> and returns components.
func parseGitPlusSSH(raw string) (*gitSSH, error) {
    if !strings.HasPrefix(raw, "git+ssh://") {
        return nil, errors.New("not git+ssh scheme")
    }
    s := strings.TrimPrefix(raw, "git+") // ssh://...
    tag := ""
    if i := strings.Index(s, "#"); i >= 0 {
        tag = s[i+1:]
        s = s[:i]
    }
    if tag == "" {
        return nil, errors.New("missing #<semver-tag> in url")
    }
    if !isSemVer(tag) {
        return nil, errors.New("invalid semver tag")
    }
    u, err := url.Parse(s)
    if err != nil {
        return nil, err
    }
    host := u.Host
    p := u.Path
    if p == "" && u.Opaque != "" {
        p = u.Opaque
    }
    if !strings.HasSuffix(p, ".git") {
        if strings.HasSuffix(p, "/") { p = p[:len(p)-1] }
        p = p + ".git"
    }
    // Ensure user is set to 'git' for ssh
    sshURL := "ssh://git@" + host + p
    repo := strings.TrimSuffix(filepath.Base(p), ".git")
    return &gitSSH{SSHURL: sshURL, Host: host, Path: p, Repo: repo, Tag: tag}, nil
}

