package driver

import (
	"testing"
)

func TestDriver_Compile_EmitsPipelinesSchema(t *testing.T) {
	// Stub file system read to return a simple source with one function and pipeline
	orig := osReadFile
	defer func() { osReadFile = orig }()
	osReadFile = func(path string) ([]byte, error) {
		src := `package main
func x(ctx Context, ev Event<T>, st State) Event<U>
pipeline P { Transform(x) }`
		return []byte(src), nil
	}
	res, err := Compile([]string{"src/main.ami"}, Options{})
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	if len(res.Pipelines) != 1 {
		t.Fatalf("expected 1 pipelines unit; got %d", len(res.Pipelines))
	}
	p := res.Pipelines[0]
	if p.Schema != "pipelines.v1" || p.Package == "" || p.File == "" {
		t.Fatalf("invalid pipelines schema header: %+v", p)
	}
	if len(p.Pipelines) != 1 || p.Pipelines[0].Name != "P" {
		t.Fatalf("unexpected pipelines content: %+v", p.Pipelines)
	}
}
