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

    git "github.com/go-git/go-git/v5"
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
