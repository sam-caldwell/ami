package root_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
	testutil "github.com/sam-caldwell/ami/src/internal/testutil"
	sch "github.com/sam-caldwell/ami/src/schemas"
)

// Assert LIFO edge is recognized and emitted in edges.json and build output.
func TestBuild_LIFOEdge_DebugArtifacts(t *testing.T) {
	t.Setenv("AMI_SEM_DIAGS", "0")
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	_, restore := testutil.ChdirToBuildTest(t)
	defer restore()

	// workspace and source
	wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: NUM_CPU
    target: ./build
    env: []
  linker: {}
  linter: {}
packages:
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
	if err := os.WriteFile("ami.workspace", []byte(wsContent), 0o644); err != nil {
		t.Fatalf("write workspace: %v", err)
	}
	if err := os.MkdirAll("src", 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	src := `package main
func f(ctx Context, ev Event<string>, st State) Event<string> { }
pipeline P { Ingress(cfg).Transform(f).Egress(in=edge.LIFO(minCapacity=2,maxCapacity=4,backpressure=block,type=int)) }
`
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// run build --verbose
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "build", "--verbose"}
	out := captureStdoutBuild(t, func() { _ = rootcmd.Execute() })
	if !strings.Contains(out, "edge_init label=P.step2.in kind=lifo") {
		t.Fatalf("expected lifo edge_init in build output; got:\n%s", out)
	}

	// edges summary JSON present and includes our LIFO edge
	epath := filepath.Join("build", "debug", "asm", "main", "edges.json")
	b, err := os.ReadFile(epath)
	if err != nil {
		t.Fatalf("missing edges.json: %v", err)
	}
	var edges sch.EdgesV1
	if err := json.Unmarshal(b, &edges); err != nil {
		t.Fatalf("unmarshal edges: %v", err)
	}
	if err := edges.Validate(); err != nil {
		t.Fatalf("edges validate: %v", err)
	}
	found := false
	for _, it := range edges.Items {
		if it.Label == "P.step2.in" && it.Kind == "edge.LIFO" && it.MinCapacity == 2 && it.MaxCapacity == 4 && it.Backpressure == "block" && it.Type == "int" && it.Bounded && it.Delivery == "atLeastOnce" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("edges.json missing expected LIFO edge; got: %+v", edges.Items)
	}
}
