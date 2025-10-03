package e2e

import (
    "context"
    "bytes"
    "encoding/json"
    "crypto/sha256"
    "encoding/hex"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "testing"
    stdtime "time"
    "github.com/sam-caldwell/ami/src/testutil"
)

type auditJSON struct {
    SumFound       bool     `json:"sumFound"`
    MissingInSum   []string `json:"missingInSum"`
    Unsatisfied    []string `json:"unsatisfied"`
    MissingInCache []string `json:"missingInCache"`
    Mismatched     []string `json:"mismatched"`
    ParseErrors    []string `json:"parseErrors"`
}

func buildAmi(t *testing.T) string {
    t.Helper()
    wd, _ := os.Getwd()
    repo := filepath.Dir(filepath.Dir(wd)) // tests/e2e -> tests -> repo root
    bin := filepath.Join(repo, "build", "ami")
    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(60*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "./src/cmd/ami")
    cmd.Dir = repo
    cmd.Env = os.Environ()
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("build ami: %v\n%s", err, string(out)) }
    return bin
}

func TestAmiModAudit_JSON_NoSum(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "audit", "json_nosum")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    yaml := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0', 'modB 1.0.0' ]\n")
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), yaml, 0o644); err != nil { t.Fatalf("write ws: %v", err) }

    // Launch `ami mod audit --json`
    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "mod", "audit", "--json")
    cmd.Dir = ws
    var stdin bytes.Buffer
    stdin.WriteString("")
    cmd.Stdin = &stdin
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String()) }
    if stderr.Len() != 0 { t.Fatalf("expected empty stderr, got: %s", stderr.String()) }

    var res auditJSON
    if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
        t.Fatalf("json: %v; stdout=%s", err, stdout.String())
    }
    if res.SumFound { t.Fatalf("sumFound true; expected false") }
    if len(res.MissingInSum) != 2 { t.Fatalf("missingInSum: %v", res.MissingInSum) }
}

func TestAmiModAudit_Human_OK(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "audit", "human_ok")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    yaml := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0' ]\n")
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), yaml, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // Prepare cache and sum
    cache := filepath.Join(ws, "cache")
    if err := os.MkdirAll(filepath.Join(cache, "modA", "1.2.3"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(cache, "modA", "1.2.3", "x.txt"), []byte("hi"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // compute sha via ami (call `ami mod sum` code path is heavy); reuse internal helper by spawning a small Go program would be overkill.
    // Instead create a matching ami.sum by running the CLI: we can simulate via workspace library compiled into the binary, but keep test blackbox.
    // We will compute hash here using a small inline function equivalent.
    sha, err := hashDirLike(cache, filepath.Join("modA", "1.2.3"))
    if err != nil { t.Fatalf("hash: %v", err) }
    sum := []byte("{\n  \"schema\": \"ami.sum/v1\",\n  \"packages\": {\n    \"modA\": {\n      \"1.2.3\": \"" + sha + "\"\n    }\n  }\n}\n")
    if err := os.WriteFile(filepath.Join(ws, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }

    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "mod", "audit")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String()) }
    if stderr.Len() != 0 { t.Fatalf("expected empty stderr, got: %s", stderr.String()) }
    if !bytes.Contains(stdout.Bytes(), []byte("ok:")) {
        t.Fatalf("expected ok summary; out=%s", stdout.String())
    }
}

// hashDirLike replicates the stable hashing used by workspace.HashDir without importing internal packages in e2e.
func hashDirLike(root string, rel string) (string, error) {
    // compute sha256 over sorted file list of root/rel
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
    // sort
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
