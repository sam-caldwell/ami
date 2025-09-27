package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

// Verify network failures return exit.Network and JSON includes a helpful message.
func TestModGet_NetworkFailure_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_get", "network_fail")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }

    // Unreachable ssh host (fast DNS failure). No #tag so tag listing occurs first and should fail as network.
    src := "git+ssh://git@invalid.invalid/org/repo.git#v0.1.0"
    var buf bytes.Buffer
    err := runModGet(&buf, dir, src, true)
    if err == nil { t.Fatalf("expected network error") }
    if exit.UnwrapCode(err) != exit.Network {
        t.Fatalf("expected exit.Network, got %v", exit.UnwrapCode(err))
    }
    var res modGetResult
    if jsonErr := json.Unmarshal(buf.Bytes(), &res); jsonErr != nil { t.Fatalf("json: %v; out=%s", jsonErr, buf.String()) }
    if res.Message == "" {
        t.Fatalf("expected message in JSON result; got: %+v", res)
    }
}

