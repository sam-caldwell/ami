package root_test

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	manifest "github.com/sam-caldwell/ami/src/ami/manifest"
	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
	testutil "github.com/sam-caldwell/ami/src/internal/testutil"
	sch "github.com/sam-caldwell/ami/src/schemas"
)

// helper to capture stdout
func captureStdoutBuild(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	w.Close()
	var b strings.Builder
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		b.WriteString(sc.Text())
		b.WriteByte('\n')
	}
	return b.String()
}

func fileSHA256Test(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func TestBuildDebugArtifacts_SchemasAndSha256(t *testing.T) {
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
import "fmt"
import (
  "foo/bar"
  alias "baz/qux"
)
pipeline P {
  Ingress(cfg).Transform(f).Egress(in=edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=block,type=[]byte))
}
func f(ctx Context, ev Event<string>, st State) Event<string> { }
`
	if err := os.WriteFile(filepath.Join("src", "main.ami"), []byte(src), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// run build --verbose
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "build", "--verbose"}
	out := captureStdoutBuild(t, func() { _ = rootcmd.Execute() })
	if !strings.Contains(out, "edge_init label=P.step2.in kind=fifo") {
		t.Fatalf("expected edge_init pseudo-op in build output; got:\n%s", out)
	}

	// resolved sources
	b, err := os.ReadFile("build/debug/source/resolved.json")
	if err != nil {
		t.Fatalf("missing resolved.json: %v", err)
	}
	var sources sch.SourcesV1
	if err := json.Unmarshal(b, &sources); err != nil {
		t.Fatalf("unmarshal sources: %v", err)
	}
	if err := sources.Validate(); err != nil {
		t.Fatalf("sources validate: %v", err)
	}
	if len(sources.Units) == 0 {
		t.Fatalf("no source units recorded")
	}
	// ensure imports present
	gotImps := sources.Units[0].Imports
	expect := map[string]bool{"fmt": true, "foo/bar": true, "baz/qux": true}
	for k := range expect {
		found := false
		for _, v := range gotImps {
			if v == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected import %s not found; got %v", k, gotImps)
		}
	}

	// asm index and file sha256
	idxBytes, err := os.ReadFile("build/debug/asm/main/index.json")
	if err != nil {
		t.Fatalf("missing asm index: %v", err)
	}
	var idx sch.ASMIndexV1
	if err := json.Unmarshal(idxBytes, &idx); err != nil {
		t.Fatalf("unmarshal asm index: %v", err)
	}
	if err := idx.Validate(); err != nil {
		t.Fatalf("asm index validate: %v", err)
	}
	if len(idx.Files) == 0 {
		t.Fatalf("asm index has no files")
	}
	// If embedded edges exist, ensure expected FIFO is present
	if len(idx.Edges) > 0 {
		has := false
		for _, it := range idx.Edges {
			if it.Label == "P.step2.in" && it.Kind == "edge.FIFO" {
				has = true
				break
			}
		}
		if !has {
			t.Fatalf("asm index 'edges' missing expected FIFO edge: %+v", idx.Edges)
		}
	}
	// Find main.ami.s entry and verify sha/size match actual file
	var asmPath string
	for _, f := range idx.Files {
		if strings.HasSuffix(f.Path, "/main.ami.s") {
			asmPath = f.Path
			break
		}
	}
	if asmPath == "" {
		t.Fatalf("did not find main.ami.s in asm index")
	}
	sha, size, err := fileSHA256Test(asmPath)
	if err != nil {
		t.Fatalf("compute sha: %v", err)
	}
	var rec sch.ASMFile
	for _, f := range idx.Files {
		if f.Path == asmPath {
			rec = f
			break
		}
	}
	if rec.Path == "" {
		t.Fatalf("asm index record not found for %s", asmPath)
	}
	if rec.Sha256 != sha || rec.Size != size {
		t.Fatalf("asm index sha/size mismatch: have sha=%s size=%d want sha=%s size=%d", rec.Sha256, rec.Size, sha, size)
	}

	// edges summary JSON present and includes our edge
	ebytes, err := os.ReadFile("build/debug/asm/main/edges.json")
	if err != nil {
		t.Fatalf("missing edges.json: %v", err)
	}
	var edges sch.EdgesV1
	if err := json.Unmarshal(ebytes, &edges); err != nil {
		t.Fatalf("unmarshal edges: %v", err)
	}
	if err := edges.Validate(); err != nil {
		t.Fatalf("edges validate: %v", err)
	}
	foundEdge := false
	for _, it := range edges.Items {
		if it.Label == "P.step2.in" && it.Kind == "edge.FIFO" && it.MinCapacity == 1 && it.MaxCapacity == 2 && it.Backpressure == "block" && it.Type == "[]byte" && it.Bounded && it.Delivery == "atLeastOnce" {
			foundEdge = true
			break
		}
	}
	if !foundEdge {
		t.Fatalf("edges.json does not include expected FIFO edge; got: %+v", edges.Items)
	}

	// eventmeta.json present and validates
	emBytes, err := os.ReadFile("build/debug/ir/main/main.ami.eventmeta.json")
	if err != nil {
		t.Fatalf("missing eventmeta.json: %v", err)
	}
	var em sch.EventMetaV1
	if err := json.Unmarshal(emBytes, &em); err != nil {
		t.Fatalf("unmarshal eventmeta: %v", err)
	}
	if err := em.Validate(); err != nil {
		t.Fatalf("eventmeta validate: %v", err)
	}
	if !em.ImmutablePayload {
		t.Fatalf("expected immutablePayload true")
	}
	if em.Trace == nil {
		t.Fatalf("expected trace context present in eventmeta")
	}
	if em.Trace.Traceparent.Name != "traceparent" || em.Trace.Traceparent.Type != "string" {
		t.Fatalf("unexpected traceparent field: %+v", em.Trace.Traceparent)
	}
	// manifest includes artifact for asm file with same sha/size
	man, err := manifest.Load("ami.manifest")
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := man.Validate(); err != nil {
		t.Fatalf("manifest validate: %v", err)
	}
	var found bool
	for _, a := range man.Artifacts {
		if a.Path == asmPath {
			found = true
			if a.Sha256 != sha || a.Size != size {
				t.Fatalf("manifest artifact sha/size mismatch: have sha=%s size=%d want sha=%s size=%d", a.Sha256, a.Size, sha, size)
			}
		}
	}
	if !found {
		t.Fatalf("manifest missing artifact for %s", asmPath)
	}
}
