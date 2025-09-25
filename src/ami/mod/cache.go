package mod

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "sort"

    git "github.com/go-git/go-git/v5"
    gitcfg "github.com/go-git/go-git/v5/config"
    "github.com/go-git/go-git/v5/plumbing"
    gogitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func CacheDir() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil { return "", err }
    dir := filepath.Join(home, ".ami", "pkg")
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    return dir, nil
}

// ami.sum structure
type Sum struct {
    Schema   string                         `json:"schema"`
    Packages map[string]map[string]string   `json:"packages"` // name -> version -> sha256
}

func loadSum(path string) (*Sum, error) {
    b, err := os.ReadFile(path)
    if err != nil { if os.IsNotExist(err) { return &Sum{Schema: "ami.sum/v1", Packages: map[string]map[string]string{}}, nil }; return nil, err }
    var s Sum
    if err := json.Unmarshal(b, &s); err != nil { return nil, err }
    if s.Schema == "" { s.Schema = "ami.sum/v1" }
    if s.Packages == nil { s.Packages = map[string]map[string]string{} }
    return &s, nil
}

func saveSum(path string, s *Sum) error {
    if s.Schema == "" { s.Schema = "ami.sum/v1" }
    if s.Packages == nil { s.Packages = map[string]map[string]string{} }
    b, err := json.MarshalIndent(s, "", "  ")
    if err != nil { return err }
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, b, 0644); err != nil { return err }
    return os.Rename(tmp, path)
}

// Expose helpers for CLI without exporting internals broadly
func LoadSumForCLI(path string) (*Sum, error) { return loadSum(path) }
func CommitDigestForCLI(repoPath, tag string) (string, error) { return commitDigest(repoPath, tag) }

// Get fetches a package given a URL (git+ssh://host/path#tag or ./localpath)
func Get(url string) (string, error) {
    cache, err := CacheDir()
    if err != nil { return "", err }
    // local path
    if strings.HasPrefix(url, "./") || strings.HasPrefix(url, "../") || strings.HasPrefix(url, "/") {
        src := filepath.Clean(url)
        name := filepath.Base(src)
        dest := filepath.Join(cache, name+"@local")
        _ = os.RemoveAll(dest)
        if err := copyDir(src, dest); err != nil { return "", err }
        return dest, nil
    }
    // git+ssh
    if strings.HasPrefix(url, "git+ssh://") {
        u := strings.TrimPrefix(url, "git+") // ssh://...
        // split tag
        tag := ""
        if i := strings.Index(u, "#"); i >= 0 {
            tag = u[i+1:]
            u = u[:i]
        }
        if tag == "" { return "", errors.New("missing #<semver-tag> in url") }
        // dest folder uses repo name and version
        base := filepath.Base(strings.TrimSuffix(u, ".git"))
        dest := filepath.Join(cache, fmt.Sprintf("%s@%s", base, tag))
        _ = os.RemoveAll(dest)
        auth, _ := gogitssh.NewSSHAgentAuth("git")
        repo, err := git.PlainClone(dest, false, &git.CloneOptions{URL: u, Auth: auth, Depth: 1})
        if err != nil { return "", err }
        wt, err := repo.Worktree()
        if err != nil { return "", err }
        // checkout tag
        if err := wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewTagReferenceName(tag)}); err != nil { return "", err }
        return dest, nil
    }
    return "", fmt.Errorf("unsupported url: %s", url)
}

func List() ([]string, error) {
    cache, err := CacheDir()
    if err != nil { return nil, err }
    entries, err := os.ReadDir(cache)
    if err != nil { return nil, err }
    out := []string{}
    for _, e := range entries { if e.IsDir() { out = append(out, e.Name()) } }
    sort.Strings(out)
    return out, nil
}

func UpdateSum(sumPath, pkg, version, repoPath, tag string) error {
    s, err := loadSum(sumPath)
    if err != nil { return err }
    if s.Packages[pkg] == nil { s.Packages[pkg] = map[string]string{} }
    digest, err := commitDigest(repoPath, tag)
    if err != nil { return err }
    s.Packages[pkg][version] = digest
    return saveSum(sumPath, s)
}

// UpdateFromWorkspace parses ami.workspace and updates dependencies to selected tags.
// Supports entries like:
//   - github.com/org/repo ==latest
//   - github.com/org/repo v1.2.3
func UpdateFromWorkspace(wsPath string) error {
    b, err := os.ReadFile(wsPath)
    if err != nil { return err }
    lines := strings.Split(string(b), "\n")
    inImport := false
    deps := [][2]string{} // [repo, constraint]
    for _, ln := range lines {
        s := strings.TrimSpace(ln)
        if strings.HasPrefix(s, "import:") { inImport = true; continue }
        if inImport {
            if strings.HasPrefix(s, "-") {
                item := strings.TrimSpace(strings.TrimPrefix(s, "-"))
                // split by spaces, first token is path, rest constraint tokens
                parts := strings.Fields(item)
                if len(parts) >= 1 {
                    repo := parts[0]
                    if strings.HasPrefix(repo, "./") { continue }
                    constraint := "==latest"
                    for _, p := range parts[1:] {
                        if p == "==latest" || strings.HasPrefix(p, "v") {
                            constraint = p
                            break
                        }
                    }
                    deps = append(deps, [2]string{repo, constraint})
                }
            } else if s == "" || strings.HasPrefix(s, "#") {
                // keep scanning
            } else if strings.HasPrefix(s, "-") == false {
                // end of import block
                inImport = false
            }
        }
    }
    // For each dep, resolve tag and fetch into cache
    for _, d := range deps {
        repoPath := d[0]
        cons := d[1]
        // construct ssh URL
        sshURL := toSSHURL(repoPath)
        tag := cons
        if cons == "==latest" {
            t, err := latestTag(sshURL)
            if err != nil { return err }
            tag = t
        }
        if !strings.HasPrefix(tag, "v") { return fmt.Errorf("invalid or unsupported version: %s", tag) }
        // Fetch using Get, then update ami.sum
        dest, err := Get("git+" + sshURL + "#" + tag)
        if err != nil { return err }
        // pkg name uses host/path
        u := strings.TrimPrefix(sshURL, "ssh://")
        u = strings.TrimPrefix(u, "git@")
        parts := strings.SplitN(u, ":", 2)
        host := parts[0]
        path := ""
        if len(parts) == 2 { path = parts[1] } else { if i:=strings.Index(host, "/"); i>0 { path=host[i+1:]; host=host[:i] } }
        pkg := filepath.Join(host, strings.TrimSuffix(path, ".git"))
        if err := UpdateSum("ami.sum", pkg, tag, dest, tag); err != nil { return err }
    }
    return nil
}

func toSSHURL(repo string) string {
    // repo like github.com/org/repo
    if strings.HasPrefix(repo, "ssh://") || strings.HasPrefix(repo, "git@") { return repo }
    return "ssh://git@" + strings.TrimSuffix(repo, ".git") + ".git"
}

// latestTag returns the latest semver tag (non-prerelease) from remote.
func latestTag(sshURL string) (string, error) {
    // Build remote and list refs
    r := git.NewRemote(nil, &gitcfg.RemoteConfig{URLs: []string{sshURL}})
    auth, _ := gogitssh.NewSSHAgentAuth("git")
    refs, err := r.List(&git.ListOptions{Auth: auth})
    if err != nil { return "", err }
    best := ""
    for _, ref := range refs {
        name := ref.Name().String()
        if strings.HasPrefix(name, "refs/tags/") {
            tag := strings.TrimPrefix(name, "refs/tags/")
            if isSemVer(tag) && !isPrerelease(tag) {
                if best == "" || semverLess(best, tag) {
                    best = tag
                }
            }
        }
    }
    if best == "" { return "", errors.New("no semver tags found") }
    return best, nil
}

func isSemVer(v string) bool { return versionRe.MatchString(v) }
func isPrerelease(v string) bool { return strings.Contains(v, "-") }

// semverLess returns true if a<b (so if best==a, and semverLess(best, tag) is true, pick tag)
func semverLess(a, b string) bool {
    pa := parseSemVer(a)
    pb := parseSemVer(b)
    if pa[0] != pb[0] { return pa[0] < pb[0] }
    if pa[1] != pb[1] { return pa[1] < pb[1] }
    return pa[2] < pb[2]
}

func parseSemVer(v string) [3]int {
    v = strings.TrimPrefix(v, "v")
    // strip prerelease/build
    if i := strings.IndexAny(v, "-+"); i >= 0 { v = v[:i] }
    parts := strings.Split(v, ".")
    out := [3]int{}
    for i := 0; i < 3 && i < len(parts); i++ {
        // ignore errors -> zero
        var n int
        for _, ch := range parts[i] { if ch < '0' || ch > '9' { n = 0; break } }
        fmt.Sscanf(parts[i], "%d", &out[i])
        _ = n
    }
    return out
}

// commitDigest computes SHA-256 of the raw commit object for the tag.
func commitDigest(repoPath, tag string) (string, error) {
    repo, err := git.PlainOpen(repoPath)
    if err != nil { return "", err }
    ref, err := repo.Reference(plumbing.NewTagReferenceName(tag), true)
    if err != nil { return "", err }
    encObj, err := repo.Storer.EncodedObject(plumbing.CommitObject, ref.Hash())
    if err != nil { return "", err }
    r, err := encObj.Reader()
    if err != nil { return "", err }
    defer r.Close()
    data, err := io.ReadAll(r)
    if err != nil { return "", err }
    header := append([]byte(fmt.Sprintf("commit %d", len(data))), 0)
    h := sha256.Sum256(append(header, data...))
    return hex.EncodeToString(h[:]), nil
}

var versionRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// copyDir copies a directory recursively.
func copyDir(src, dst string) error {
    return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
        if err != nil { return err }
        rel, _ := filepath.Rel(src, p)
        tgt := filepath.Join(dst, rel)
        if info.IsDir() {
            return os.MkdirAll(tgt, info.Mode())
        }
        b, err := os.ReadFile(p)
        if err != nil { return err }
        return os.WriteFile(tgt, b, info.Mode())
    })
}

// Expose helpers for CLI without exporting internals broadly


// resolveConstraint selects a tag for a given constraint string.
func resolveConstraint(sshURL, cons string) (string, error) {
    cons = strings.TrimSpace(cons)
    if cons == "" || cons == "==latest" {
        return latestTag(sshURL)
    }
    if isSemVer(cons) {
        return cons, nil
    }
    r := git.NewRemote(nil, &gitcfg.RemoteConfig{URLs: []string{sshURL}})
    auth, _ := gogitssh.NewSSHAgentAuth("git")
    refs, err := r.List(&git.ListOptions{Auth: auth})
    if err != nil { return "", err }
    tags := collectSemverTags(refs)
    if len(tags)==0 { return "", errors.New("no semver tags found") }
    sortSemver(tags)
    var filt []string
    switch {
    case strings.HasPrefix(cons, ">="):
        base := strings.TrimSpace(strings.TrimPrefix(cons, ">="))
        for _, t := range tags { if !semverLess(t, base) { filt = append(filt, t) } }
    case strings.HasPrefix(cons, ">"):
        base := strings.TrimSpace(strings.TrimPrefix(cons, ">"))
        for _, t := range tags { if semverLess(base, t) { filt = append(filt, t) } }
    case strings.HasPrefix(cons, "^"):
        base := strings.TrimSpace(strings.TrimPrefix(cons, "^"))
        bv := parseSemVer(base)
        for _, t := range tags { tv:=parseSemVer(t); if tv[0]==bv[0] && !semverLess(t, base) { filt=append(filt,t) } }
    case strings.HasPrefix(cons, "~"):
        base := strings.TrimSpace(strings.TrimPrefix(cons, "~"))
        bv := parseSemVer(base)
        for _, t := range tags { tv:=parseSemVer(t); if tv[0]==bv[0] && tv[1]==bv[1] && !semverLess(t, base) { filt=append(filt,t) } }
    default:
        return "", fmt.Errorf("unsupported constraint: %s", cons)
    }
    if len(filt)==0 { return "", errors.New("no matching tags for constraint") }
    return filt[len(filt)-1], nil
}

func collectSemverTags(refs []*plumbing.Reference) []string {
    out := []string{}
    for _, ref := range refs {
        n := ref.Name().String()
        if strings.HasPrefix(n, "refs/tags/") {
            tag := strings.TrimPrefix(n, "refs/tags/")
            if isSemVer(tag) { out = append(out, tag) }
        }
    }
    return out
}

func sortSemver(tags []string) {
    sort.Slice(tags, func(i,j int) bool { return semverLess(tags[i], tags[j]) })
}
