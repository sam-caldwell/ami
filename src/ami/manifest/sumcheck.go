package manifest

import (
    "encoding/json"
    "errors"
    "os"
)

// CrossCheckWithSumFile loads ami.sum at sumPath and ensures each package
// listed in the manifest matches a version→digest entry.
func (m *Manifest) CrossCheckWithSumFile(sumPath string) error {
    if m == nil {
        return errors.New("nil manifest")
    }
    b, err := os.ReadFile(sumPath)
    if err != nil {
        return err
    }
    // Minimal shape for ami.sum
    var sum struct {
        Schema   string                       `json:"schema"`
        Packages map[string]map[string]string `json:"packages"`
    }
    if err := json.Unmarshal(b, &sum); err != nil {
        return err
    }
    for _, p := range m.Packages {
        vers, ok := sum.Packages[p.Name]
        if !ok {
            return errors.New("ami.sum missing package: " + p.Name)
        }
        d, ok := vers[p.Version]
        if !ok {
            return errors.New("ami.sum missing version for package: " + p.Name)
        }
        if d != p.Digest {
            return errors.New("ami.sum digest mismatch for package: " + p.Name)
        }
    }
    return nil
}

