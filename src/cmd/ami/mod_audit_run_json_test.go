package main

import (
    "bytes"
    "encoding/json"
    "crypto/sha256"
    "encoding/hex"
    "os"
    "path/filepath"
    "sort"
    "testing"
)

// Verify runModAudit emits expected JSON fields when no ami.sum is present.
func TestModAudit_Run_JSON_NoSum(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_mod_audit", "json_nosum")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0', 'modB 1.0.0' ]\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }

    var out bytes.Buffer
    if err := runModAudit(&out, dir, true); err != nil { t.Fatalf("run: %v", err) }
    var obj struct{
        SumFound       bool     `json:"sumFound"`
        MissingInSum   []string `json:"missingInSum"`
        Unsatisfied    []string `json:"unsatisfied"`
        MissingInCache []string `json:"missingInCache"`
        Mismatched     []string `json:"mismatched"`
        ParseErrors    []string `json:"parseErrors"`
        Timestamp      string   `json:"timestamp"`
    }
    if err := json.Unmarshal(out.Bytes(), &obj); err != nil {
        t.Fatalf("json: %v; out=%s", err, out.String())
    }
    if obj.SumFound { t.Fatalf("sumFound true; expected false") }
    if len(obj.MissingInSum) != 2 { t.Fatalf("missingInSum: %v", obj.MissingInSum) }
}

// hashDirLike replicates the stable hashing used by workspace hashing used in E2E tests.
func hashDirLike(root string, rel string) (string, error) {
    dir := filepath.Join(root, rel)
    var files []string
    if err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
        if err != nil { return err }
        if info.IsDir() { return nil }
        r, err := filepath.Rel(dir, p)
        if err != nil { return err }
        files = append(files, r)
        return nil
    }); err != nil { return "", err }
    sort.Strings(files)
    h := sha256.New()
    for _, f := range files {
        b, err := os.ReadFile(filepath.Join(dir, f))
        if err != nil { return "", err }
        _, _ = h.Write([]byte(f))
        _, _ = h.Write(b)
    }
    return hex.EncodeToString(h.Sum(nil)), nil
}

