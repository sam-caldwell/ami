package e2e

import (
    "bytes"
    "encoding/json"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

func stageGoModule(t *testing.T, root string, tests string) {
    t.Helper()
    if err := os.MkdirAll(root, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    if err := os.WriteFile(filepath.Join(root, "tmp_test.go"), []byte(tests), 0o644); err != nil { t.Fatalf("tests: %v", err) }
}

func TestAmiTest_Human_OK(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "ami_test", "human_ok")
    _ = os.RemoveAll(ws)
    stageGoModule(t, ws, "package tmp\nimport \"testing\"\nfunc TestA(t *testing.T){}\n")

    cmd := exec.Command(bin, "test")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    if stderr.Len() != 0 { t.Fatalf("expected empty stderr; got: %s", stderr.String()) }
    if !strings.Contains(stdout.String(), "test: OK") {
        t.Fatalf("expected 'test: OK' in stdout; got: %s", stdout.String())
    }
}

func TestAmiTest_JSON_StreamsAndSummary(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "ami_test", "json_ok")
    _ = os.RemoveAll(ws)
    stageGoModule(t, ws, "package tmp\nimport \"testing\"\nfunc TestA(t *testing.T){}\n")

    cmd := exec.Command(bin, "test", "--json")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    if stderr.Len() != 0 { t.Fatalf("expected empty stderr; got: %s", stderr.String()) }
    s := stdout.String()
    if !strings.Contains(s, "\"Action\":\"run\"") {
        t.Fatalf("expected streamed events; stdout=%s", s)
    }
    // Parse last JSON line as summary and check counts
    lines := bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte("\n"))
    var last map[string]any
    if err := json.Unmarshal(lines[len(lines)-1], &last); err != nil { t.Fatalf("summary json: %v", err) }
    if ok, _ := last["ok"].(bool); !ok { t.Fatalf("expected ok=true summary; last=%v", last) }
    if int(last["packages"].(float64)) != 1 || int(last["tests"].(float64)) != 1 || int(last["failures"].(float64)) != 0 {
        t.Fatalf("unexpected counts: %v", last)
    }
}

func TestAmiTest_Verbose_Artifacts(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "ami_test", "verbose")
    _ = os.RemoveAll(ws)
    stageGoModule(t, ws, "package tmp\nimport \"testing\"\nfunc TestA(t *testing.T){}\nfunc TestB(t *testing.T){}\n")

    cmd := exec.Command(bin, "test", "--verbose")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    if _, err := os.Stat(filepath.Join(ws, "build", "test", "test.log")); err != nil { t.Fatalf("missing test.log: %v", err) }
    b, err := os.ReadFile(filepath.Join(ws, "build", "test", "test.manifest"))
    if err != nil { t.Fatalf("read manifest: %v", err) }
    s := string(b)
    if !strings.Contains(s, "example.com/tmp TestA") || !strings.Contains(s, "example.com/tmp TestB") {
        t.Fatalf("manifest missing tests: %s", s)
    }
}

func TestAmiTest_Failing_EmitsError(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "ami_test", "fail")
    _ = os.RemoveAll(ws)
    stageGoModule(t, ws, "package tmp\nimport \"testing\"\nfunc TestFail(t *testing.T){ t.Fatal(\"boom\") }\n")

    cmd := exec.Command(bin, "test")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err == nil {
        t.Fatalf("expected non-nil error from failing tests\nstdout=%s\nstderr=%s", stdout.String(), stderr.String())
    }
}

func TestAmiTest_JSON_Failing_SummaryEmitted(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "ami_test", "json_fail")
    _ = os.RemoveAll(ws)
    stageGoModule(t, ws, "package tmp\nimport \"testing\"\nfunc TestFail(t *testing.T){ t.Fatal(\"boom\") }\n")

    cmd := exec.Command(bin, "test", "--json")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err == nil {
        t.Fatalf("expected non-zero exit for failing tests\nstdout=%s\nstderr=%s", stdout.String(), stderr.String())
    }
    // Summary should be last line in stdout and show ok:false and failures>=1
    lines := bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte("\n"))
    var last map[string]any
    if err := json.Unmarshal(lines[len(lines)-1], &last); err != nil { t.Fatalf("summary json: %v\nstdout=%s", err, stdout.String()) }
    if ok, _ := last["ok"].(bool); ok { t.Fatalf("expected ok=false; last=%v", last) }
    if int(last["failures"].(float64)) < 1 { t.Fatalf("expected failures>=1; last=%v", last) }
}
