package ascii

import (
	"testing"

	"github.com/sam-caldwell/ami/src/schemas/graph"
)

func testRenderLine_StraightChain(t *testing.T) {
	g := graph.Graph{
		Package: "main",
		Name:    "Line",
		Nodes: []graph.Node{
			{ID: "ingress", Kind: "ingress", Label: "ingress"},
			{ID: "worker", Kind: "worker", Label: "worker"},
			{ID: "egress", Kind: "egress", Label: "egress"},
		},
		Edges: []graph.Edge{{From: "ingress", To: "worker"}, {From: "worker", To: "egress"}},
	}
	got := RenderLine(g, Options{})
	want := "[ingress] --> (worker) --> [egress]"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func testRenderLine_FallbackSorted(t *testing.T) {
	g := graph.Graph{Nodes: []graph.Node{{ID: "b", Kind: "worker", Label: "B"}, {ID: "a", Kind: "ingress", Label: "A"}}}
	got := RenderLine(g, Options{})
	// Fallback sorts by ID: a then b; ingress uses [A], worker uses (B)
	want := "[A] --> (B)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
