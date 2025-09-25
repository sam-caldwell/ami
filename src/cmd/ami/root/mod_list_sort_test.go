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

type sortDiag struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// Verify `ami mod list` yields sorted entries to ensure deterministic output.
func TestModList_JSON_SortedOrder(t *testing.T) {
	// Prepare isolated HOME cache with out-of-order entries
	home := t.TempDir()
	t.Setenv("HOME", home)
	cache := filepath.Join(home, ".ami", "pkg")
	if err := os.MkdirAll(filepath.Join(cache, "b@v1.0.0"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(cache, "a@v1.0.0"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(cache, "c@v1.0.0"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Run from empty workspace (no ami.sum needed for order test)
	ws := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(ws); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "--json", "mod", "list"}

	rOut, wOut, _ := os.Pipe()
	oldOut := os.Stdout
	os.Stdout = wOut
	_ = rootcmd.Execute()
	_ = wOut.Close()
	os.Stdout = oldOut

	// Read JSON lines and collect entries in order
	var entries []string
	sc := bufio.NewScanner(rOut)
	for sc.Scan() {
		var rec sortDiag
		if err := json.Unmarshal([]byte(strings.TrimSpace(sc.Text())), &rec); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		if rec.Message != "cache.entry" {
			continue
		}
		if e, ok := rec.Data["entry"].(string); ok {
			entries = append(entries, e)
		}
	}
	want := []string{"a@v1.0.0", "b@v1.0.0", "c@v1.0.0"}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d: %v", len(entries), len(want), entries)
	}
	for i := range want {
		if entries[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %q want %q (all=%v)", i, entries[i], want[i], entries)
		}
	}
}
