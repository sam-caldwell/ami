package driver

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testCompile_ContainerLiteral_AssignAndVarInit_TypesInIR(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	code := "package app\nfunc F(){ var a slice<int>; a = slice<int>{1,2}; var m map<string,int>; m = map<string,int>{\"k\": 3} }\n"
	fs.AddFile("unit1.ami", code)
	pkgs := []Package{{Name: "app", Files: fs}}
	arts, diags := Compile(ws, pkgs, Options{Debug: true})
	if len(diags) != 0 {
		t.Fatalf("diags: %+v", diags)
	}
	if len(arts.IR) != 1 {
		t.Fatalf("expected 1 IR artifact")
	}
	b, err := os.ReadFile(arts.IR[0])
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatalf("json: %v", err)
	}
	fns := obj["functions"].([]any)
	fn := fns[0].(map[string]any)
	blks := fn["blocks"].([]any)
	blk := blks[0].(map[string]any)
	instrs := blk["instrs"].([]any)
	if len(instrs) < 4 {
		t.Fatalf("expected several instrs, got %d", len(instrs))
	}
	// Find first ASSIGN and confirm src type slice<int>
	foundSlice := false
	foundMap := false
	for _, it := range instrs {
		m := it.(map[string]any)
		switch m["op"] {
		case "ASSIGN":
			src := m["src"].(map[string]any)
			if src["type"] == "slice<int>" {
				foundSlice = true
			}
			if src["type"] == "map<string,int>" {
				foundMap = true
			}
		case "VAR":
			if init, ok := m["init"].(map[string]any); ok {
				if init["type"] == "slice<int>" {
					foundSlice = true
				}
				if init["type"] == "map<string,int>" {
					foundMap = true
				}
			}
		}
	}
	if !foundSlice {
		t.Fatalf("slice literal type not found in IR")
	}
	if !foundMap {
		t.Fatalf("map literal type not found in IR")
	}
}

func testCompile_ContainerLiteral_Set_TypeInIR(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	code := "package app\nfunc F(){ var s set<string>; s = set<string>{\"x\",\"y\"} }\n"
	fs.AddFile("unit2.ami", code)
	pkgs := []Package{{Name: "app", Files: fs}}
	arts, diags := Compile(ws, pkgs, Options{Debug: true})
	if len(diags) != 0 {
		t.Fatalf("diags: %+v", diags)
	}
	if len(arts.IR) != 1 {
		t.Fatalf("expected 1 IR artifact")
	}
	b, err := os.ReadFile(arts.IR[0])
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatalf("json: %v", err)
	}
	fns := obj["functions"].([]any)
	fn := fns[0].(map[string]any)
	blks := fn["blocks"].([]any)
	blk := blks[0].(map[string]any)
	instrs := blk["instrs"].([]any)
	found := false
	for _, it := range instrs {
		m := it.(map[string]any)
		if m["op"] == "ASSIGN" {
			src := m["src"].(map[string]any)
			if src["type"] == "set<string>" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("set literal type not found in IR")
	}
}
