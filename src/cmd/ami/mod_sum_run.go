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
)

type modSumPkg struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Sha256  string `json:"sha256"`
}

type modSumResult struct {
    Path         string   `json:"path,omitempty"`
    Ok           bool     `json:"ok"`
    PackagesSeen int      `json:"packages"`
    Schema       string   `json:"schema"`
    Verified     []string `json:"verified,omitempty"`
    Missing      []string `json:"missing,omitempty"`
    Mismatched   []string `json:"mismatched,omitempty"`
    Message      string   `json:"message,omitempty"`
}

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
    for _, p := range pkgs {
        cp := filepath.Join(cache, p.Name, p.Version)
        if st, err := os.Stat(cp); err != nil || !st.IsDir() {
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
        } else {
            mismatched = append(mismatched, key(p.Name, p.Version))
        }
    }
    sort.Strings(verified)
    sort.Strings(missing)
    sort.Strings(mismatched)
    res.Verified = verified
    res.Missing = missing
    res.Mismatched = mismatched
    res.Ok = len(missing) == 0 && len(mismatched) == 0
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

func strOrEmpty(v any) string { if s, ok := v.(string); ok { return s }; return "" }
func key(name, ver string) string { if ver == "" { return name }; return name + "@" + ver }

func hashDir(root string) (string, error) {
    h := sha256.New()
    var files []string
    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if d.IsDir() { return nil }
        rel, err := filepath.Rel(root, path)
        if err != nil { return err }
        files = append(files, rel)
        return nil
    })
    if err != nil { return "", err }
    sort.Strings(files)
    for _, rel := range files {
        p := filepath.Join(root, rel)
        b, err := os.ReadFile(p)
        if err != nil { return "", err }
        _, _ = h.Write([]byte(rel))
        _, _ = h.Write(b)
    }
    return hex.EncodeToString(h.Sum(nil)), nil
}

func equalSHA(a, b string) bool {
    if len(a) != len(b) { return false }
    var diff byte
    for i := 0; i < len(a); i++ { diff |= a[i] ^ b[i] }
    return diff == 0
}
