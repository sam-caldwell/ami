package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "sort"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "os/exec"
)

func runModSum(out io.Writer, dir string, jsonOut bool) error {
    path := filepath.Join(dir, "ami.sum")
    res := modSumResult{Path: path}
    b, err := os.ReadFile(path)
    if err != nil {
        if jsonOut { res.Message = "ami.sum not found"; _ = json.NewEncoder(out).Encode(res) }
        return fmt.Errorf("read ami.sum: %w", err)
    }
    var m map[string]any
    if err := json.Unmarshal(b, &m); err != nil {
        if jsonOut { res.Message = "invalid JSON"; _ = json.NewEncoder(out).Encode(res) }
        return fmt.Errorf("invalid ami.sum: %w", err)
    }
    schema, _ := m["schema"].(string)
    if schema != "ami.sum/v1" {
        if jsonOut { res.Schema = schema; res.Message = "unsupported schema"; _ = json.NewEncoder(out).Encode(res) }
        return fmt.Errorf("unsupported schema: %s", schema)
    }
    res.Schema = schema
    // Decode packages (support object and array form)
    var pkgs []modSumPkg
    if p, ok := m["packages"]; ok {
        switch t := p.(type) {
        case []any:
            for _, el := range t {
                if mm, ok := el.(map[string]any); ok {
                    pkgs = append(pkgs, modSumPkg{
                        Name:    strOrEmpty(mm["name"]),
                        Version: strOrEmpty(mm["version"]),
                        Sha256:  strOrEmpty(mm["sha256"]),
                        Source:  strOrEmpty(mm["source"]),
                    })
                }
            }
        case map[string]any:
            for name, v := range t {
                if mm, ok := v.(map[string]any); ok {
                    pkgs = append(pkgs, modSumPkg{
                        Name:    name,
                        Version: strOrEmpty(mm["version"]),
                        Sha256:  strOrEmpty(mm["sha256"]),
                        Source:  strOrEmpty(mm["source"]),
                    })
                }
            }
        }
    }
    res.PackagesSeen = len(pkgs)

    // Check integrity vs cache
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache == "" {
        home, _ := os.UserHomeDir()
        cache = filepath.Join(home, ".ami", "pkg")
    }
    var verified, missing, mismatched []string
    present := make(map[string]struct{})
    updatedSum := false
    for _, p := range pkgs {
        cp := filepath.Join(cache, p.Name, p.Version)
        if st, err := os.Stat(cp); err != nil || !st.IsDir() {
            // Attempt git fetch if source provided
            if isGitSource(p.Source) && p.Version != "" {
                if err := fetchGitToCache(p.Source, p.Version, cp); err != nil {
                    missing = append(missing, key(p.Name, p.Version))
                    continue
                }
                // After fetching, compute directory hash and update ami.sum entry; also attach commit digest if available
                got, herr := hashDir(cp)
                if herr != nil {
                    mismatched = append(mismatched, key(p.Name, p.Version))
                    continue
                }
                // Try to compute commit digest for traceability; ignore errors
                commitDigest, _ := computeCommitDigest(cp, p.Version)
                if !equalSHA(got, p.Sha256) {
                    // Update m in-place (sha256 and optional commit)
                    if updateSumEntryWithCommit(m, p.Name, p.Version, got, commitDigest, p.Source) {
                        updatedSum = true
                    }
                } else if commitDigest != "" {
                    // Ensure commit field present even when sha matches
                    if updateSumEntryWithCommit(m, p.Name, p.Version, p.Sha256, commitDigest, p.Source) {
                        updatedSum = true
                    }
                }
                verified = append(verified, key(p.Name, p.Version))
                present[key(p.Name, p.Version)] = struct{}{}
                continue
            }
            // no source; cannot auto-fetch
            missing = append(missing, key(p.Name, p.Version))
            continue
        }
        got, err := hashDir(cp)
        if err != nil {
            mismatched = append(mismatched, key(p.Name, p.Version))
            continue
        }
        if equalSHA(got, p.Sha256) {
            verified = append(verified, key(p.Name, p.Version))
            present[key(p.Name, p.Version)] = struct{}{}
        } else {
            mismatched = append(mismatched, key(p.Name, p.Version))
        }
    }
    // Cross-check workspace declared packages against ami.sum entries.
    // If a workspace package with name+version exists but is not present in ami.sum,
    // flag it as missing to prompt users to run `ami mod get` or update sum.
    var ws workspace.Workspace
    if err := ws.Load(filepath.Join(dir, "ami.workspace")); err == nil {
        for _, e := range ws.Packages {
            name := e.Package.Name
            ver := e.Package.Version
            if name == "" || ver == "" { continue }
            k := key(name, ver)
            if _, ok := present[k]; !ok {
                // Only add if cache holds the package or sum truly lacks it?
                // We treat absence from sum as missing regardless of cache state.
                missing = append(missing, k)
            }
        }
    }
    sort.Strings(verified)
    sort.Strings(missing)
    sort.Strings(mismatched)
    res.Verified = verified
    res.Missing = missing
    res.Mismatched = mismatched
    res.Ok = len(missing) == 0 && len(mismatched) == 0
    // If sum updated, write back
    if updatedSum {
        if b, err := json.MarshalIndent(m, "", "  "); err == nil {
            _ = os.WriteFile(path, b, 0o644)
        }
    }
    if jsonOut {
        if !res.Ok {
            res.Message = "integrity failure"
            _ = json.NewEncoder(out).Encode(res)
            return exit.New(exit.Integrity, "integrity failure")
        }
        res.Message = "ok"
        return json.NewEncoder(out).Encode(res)
    }
    if res.Ok {
        _, _ = io.WriteString(out, "ok\n")
        return nil
    }
    return exit.New(exit.Integrity, "integrity failure")
}

 

 

func equalSHA(a, b string) bool {
    if len(a) != len(b) { return false }
    var diff byte
    for i := 0; i < len(a); i++ { diff |= a[i] ^ b[i] }
    return diff == 0
}

func isGitSource(s string) bool {
    return len(s) > 0 && (hasPrefix(s, "git+ssh://") || hasPrefix(s, "file+git://"))
}

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }

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

// updateSumEntry updates the ami.sum map with the provided hash (and source retained if present).
func updateSumEntry(m map[string]any, name, version, sha, source string) bool {
    p, ok := m["packages"]
    if !ok { return false }
    switch t := p.(type) {
    case []any:
        for _, el := range t {
            if mm, ok := el.(map[string]any); ok {
                if strOrEmpty(mm["name"]) == name && strOrEmpty(mm["version"]) == version {
                    mm["sha256"] = sha
                    if source != "" { mm["source"] = source }
                    return true
                }
            }
        }
    case map[string]any:
        if mm, ok := t[name].(map[string]any); ok {
            mm["version"] = version
            mm["sha256"] = sha
            if source != "" { mm["source"] = source }
            t[name] = mm
            m["packages"] = t
            return true
        }
    }
    return false
}

// updateSumEntryWithCommit updates the ami.sum map with sha and optional commit digest.
func updateSumEntryWithCommit(m map[string]any, name, version, sha, commit, source string) bool {
    p, ok := m["packages"]
    if !ok { return false }
    switch t := p.(type) {
    case []any:
        for _, el := range t {
            if mm, ok := el.(map[string]any); ok {
                if strOrEmpty(mm["name"]) == name && strOrEmpty(mm["version"]) == version {
                    mm["sha256"] = sha
                    if source != "" { mm["source"] = source }
                    if commit != "" { mm["commit"] = commit }
                    return true
                }
            }
        }
    case map[string]any:
        if mm, ok := t[name].(map[string]any); ok {
            mm["version"] = version
            mm["sha256"] = sha
            if source != "" { mm["source"] = source }
            if commit != "" { mm["commit"] = commit }
            t[name] = mm
            m["packages"] = t
            return true
        }
    }
    return false
}
