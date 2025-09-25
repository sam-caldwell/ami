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

func TestLint_JSON_Warns_On_UnboundedBlockEdge_And_NoWorkers(t *testing.T) {
	ws := t.TempDir()
	content := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	// Pipeline with Transform() (no workers) and an edge with block backpressure and unbounded capacity
	src := "package main\npipeline P { Ingress(cfg).Transform().Egress(cfg, in=edge.FIFO(backpressure=\"block\", maxCapacity=0)) }\n"
	if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	oldDir, _ := os.Getwd()
	_ = os.Chdir(ws)
	defer os.Chdir(oldDir)
	oldArgs := os.Args
	out := captureStdoutLintRules(t, func() { os.Args = []string{"ami", "--json", "lint"}; _ = rootcmd.Execute() })
	os.Args = oldArgs
	seenEdge := false
	seenNoWorkers := false
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var m map[string]any
		if json.Unmarshal([]byte(sc.Text()), &m) != nil {
			continue
		}
		if m["schema"] != "diag.v1" {
			continue
		}
		switch m["code"] {
		case "W_EDGE_SMELL_UNBOUNDED_BLOCK":
			seenEdge = true
		case "W_NODE_NO_WORKERS":
			seenNoWorkers = true
		}
	}
	if !seenEdge || !seenNoWorkers {
		t.Fatalf("expected both W_EDGE_SMELL_UNBOUNDED_BLOCK and W_NODE_NO_WORKERS; out=\n%s", out)
	}
}
