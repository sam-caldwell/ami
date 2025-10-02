package sem

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

func wantCode(t *testing.T, ds []diag.Record, code string) {
	t.Helper()
	for _, d := range ds {
		if d.Code == code {
			return
		}
	}
	t.Fatalf("expected diag %s, got %+v", code, ds)
}

func testPipelineSemantics_DisconnectedNode_WhenEdgesPresent(t *testing.T) {
	// B is declared but has no edges; edges are present elsewhere.
	src := "package app\npipeline P(){ ingress; A; B; egress; A -> egress; }\n"
	f := (&source.FileSet{}).AddFile("p.ami", src)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzePipelineSemantics(af)
	wantCode(t, ds, "E_PIPELINE_NODE_DISCONNECTED")
}

func testPipelineSemantics_MissingIngressToEgressPath(t *testing.T) {
	// Edges exist between A and B, but ingress/egress not connected.
	src := "package app\npipeline P(){ ingress; A; B; egress; A -> B; }\n"
	f := (&source.FileSet{}).AddFile("p.ami", src)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzePipelineSemantics(af)
	wantCode(t, ds, "E_PIPELINE_NO_PATH_INGRESS_EGRESS")
}
