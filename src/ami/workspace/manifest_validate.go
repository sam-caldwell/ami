package workspace

import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
)

// Validate checks AMI_PACKAGE_CACHE to ensure that every package@version exists and matches the recorded sha256.
// Returns slices of keys ("name@version") for verified, missing, and mismatched.
func (m *Manifest) Validate() (verified, missing, mismatched []string, err error) {
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache == "" {
        home, herr := os.UserHomeDir()
        if herr != nil { return nil, nil, nil, fmt.Errorf("resolve cache: %v", herr) }
        cache = filepath.Join(home, ".ami", "pkg")
    }
    for name, versions := range m.Packages {
        for ver, sha := range versions {
            p := filepath.Join(cache, name, ver)
            st, statErr := os.Stat(p)
            if statErr != nil || !st.IsDir() {
                missing = append(missing, name+"@"+ver)
                continue
            }
            got, hErr := HashDir(p)
            if hErr != nil { mismatched = append(mismatched, name+"@"+ver); continue }
            if got == sha { verified = append(verified, name+"@"+ver) } else { mismatched = append(mismatched, name+"@"+ver) }
        }
    }
    sort.Strings(verified)
    sort.Strings(missing)
    sort.Strings(mismatched)
    return verified, missing, mismatched, nil
}

