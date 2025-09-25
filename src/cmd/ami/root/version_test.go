package root_test

import (
    "encoding/json"
    "os"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

func TestVersion_JSON_PrintsSimpleObject(t *testing.T) {
    oldArgs := os.Args
    defer func(){ os.Args = oldArgs }()
    os.Args = []string{"ami", "--json", "version"}
    out := captureStdout(t, func(){ _ = rootcmd.Execute() })
    var obj map[string]string
    if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &obj); err != nil {
        t.Fatalf("invalid json: %v out=%q", err, out)
    }
    if v := obj["version"]; v == "" {
        t.Fatalf("missing version field: %v", obj)
    }
}

func TestVersion_Human_ContainsVersionString(t *testing.T) {
    oldArgs := os.Args
    defer func(){ os.Args = oldArgs }()
    os.Args = []string{"ami", "version"}
    out := captureStdout(t, func(){ _ = rootcmd.Execute() })
    if !strings.Contains(out, "version:") {
        t.Fatalf("expected human output to contain 'version:' got: %q", out)
    }
}
