package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// diagRecordLike is the minimal shape we need from diag.Record for tests
type diagRecordLike struct{ Code string }

func codes(ds []diag.Record) []diagRecordLike {
    out := make([]diagRecordLike, len(ds))
    for i, d := range ds { out[i] = diagRecordLike{Code: d.Code} }
    return out
}

// Ensure A->B references declared steps and forbid inbound to ingress / outbound from egress.
func TestPipelineSemantics_EdgeEndpointAndDirection_Validations(t *testing.T) {
    cases := []struct{
        name string
        src  string
        want []string
    }{
        {
            name: "undeclared to",
            src:  "package app\npipeline P(){ ingress; A; egress; A -> B; }\n",
            want: []string{"E_EDGE_UNDECLARED_TO"},
        },
        {
            name: "undeclared from",
            src:  "package app\npipeline P(){ ingress; A; egress; X -> A; }\n",
            want: []string{"E_EDGE_UNDECLARED_FROM"},
        },
        {
            name: "to ingress forbidden",
            src:  "package app\npipeline P(){ ingress; A; egress; A -> ingress; }\n",
            want: []string{"E_EDGE_TO_INGRESS"},
        },
        {
            name: "from egress forbidden",
            src:  "package app\npipeline P(){ ingress; A; egress; egress -> A; }\n",
            want: []string{"E_EDGE_FROM_EGRESS"},
        },
    }
    for _, tc := range cases {
        f := (&source.FileSet{}).AddFile("p.ami", tc.src)
        p := parser.New(f)
        af, _ := p.ParseFile()
        ds := AnalyzePipelineSemantics(af)
        got := codes(ds)
        for _, w := range tc.want {
            ok := false
            for _, g := range got { if g.Code == w { ok = true; break } }
            if !ok {
                t.Fatalf("%s: expected code %s, got %+v", tc.name, w, got)
            }
        }
    }
}
