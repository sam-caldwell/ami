package root_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

// record shape for parsing logger output
type mlRecord struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func TestModList_JSON_IncludesManifestInfoWhenAvailable(t *testing.T) {
	// Workspace
	ws := t.TempDir()
	// HOME with cache
	home := t.TempDir()
	t.Setenv("HOME", home)
	cache := filepath.Join(home, ".ami", "pkg")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	// Cache entry
	base := "repo"
	ver := "v1.2.3"
	entry := filepath.Join(cache, base+"@"+ver)
	if err := os.MkdirAll(entry, 0o755); err != nil {
		t.Fatalf("mkdir entry: %v", err)
	}

	// ami.sum with full package name and digest
	full := "github.com/example/" + base
	digest := "cafefeeddeadbeef"
	sum := `{"schema":"ami.sum/v1","packages":{"` + full + `":{"` + ver + `":"` + digest + `"}}}`
	if err := os.WriteFile(filepath.Join(ws, "ami.sum"), []byte(sum), 0o644); err != nil {
		t.Fatalf("write ami.sum: %v", err)
	}

	// ami.manifest including same package with source path
	man := `{
  "schema": "ami.manifest/v1",
  "project": {"name": "demo", "version": "0.0.1"},
  "packages": [
    {"name": "` + full + `", "version": "` + ver + `", "digestSHA256": "` + digest + `", "source": "` + entry + `"}
  ],
  "artifacts": [],
  "toolchain": {"amiVersion": "v0.0.0-dev", "goVersion": "1.25"},
  "createdAt": "2025-01-01T00:00:00.000Z"
}`
	if err := os.WriteFile(filepath.Join(ws, "ami.manifest"), []byte(man), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Run from workspace
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(ws); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Capture stdout only (JSON mode)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "--json", "mod", "list"}

	rOut, wOut, _ := os.Pipe()
	oldOut := os.Stdout
	os.Stdout = wOut
	_ = rootcmd.Execute()
	_ = wOut.Close()
	os.Stdout = oldOut

	// Parse lines
	var seen bool
	sc := bufio.NewScanner(rOut)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec mlRecord
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if rec.Message != "cache.entry" {
			continue
		}
		if rec.Data["entry"] != base+"@"+ver {
			continue
		}
		if rec.Data["manifest"].(bool) != true {
			t.Fatalf("expected manifest=true")
		}
		if rec.Data["manifestName"].(string) != full {
			t.Fatalf("manifestName mismatch: %v", rec.Data["manifestName"])
		}
		if rec.Data["manifestDigest"].(string) != digest {
			t.Fatalf("manifestDigest mismatch: %v", rec.Data["manifestDigest"])
		}
		if rec.Data["manifestSource"].(string) != entry {
			t.Fatalf("manifestSource mismatch: %v", rec.Data["manifestSource"])
		}
		seen = true
	}
	if !seen {
		t.Fatalf("did not see expected cache.entry record")
	}
}
