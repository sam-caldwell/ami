package root_test

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTest_JSON_AMIRuntime_ErrorAndTimeout(t *testing.T) {
	ws := t.TempDir()
	gomod := "module example.com/ami-test\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	wsContent := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: NUM_CPU, target: ./build, env: [] }, linker: {}, linter: {} }
packages: [ { main: { version: 0.0.1, root: ./src, import: [] } } ]
`
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil {
		t.Fatalf("write workspace: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	a := `#pragma test:case ErrOk
#pragma test:runtime pipeline=P input={"error_code":"E_OOPS"} expect_error=E_OOPS
#pragma test:case TimeoutOk
#pragma test:runtime pipeline=P input={"sleep_ms":20} timeout=5 expect_error=E_TIMEOUT
package main
Pipeline(P) { }
`
	if err := os.WriteFile(filepath.Join(ws, "src", "rt2_test.ami"), []byte(a), 0o644); err != nil {
		t.Fatalf("write rt2: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON=1")
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			_ = ee
		} else {
			t.Fatalf("unexpected error: %v\n%s", err, string(out))
		}
	}
	// Verify both cases passed
	want := map[string]bool{"ErrOk": false, "TimeoutOk": false}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		var obj map[string]any
		if json.Unmarshal([]byte(sc.Text()), &obj) != nil {
			continue
		}
		if obj["schema"] == "test.v1" && obj["type"] == "test_end" && obj["status"] == "pass" {
			name, _ := obj["name"].(string)
			if _, ok := want[name]; ok {
				want[name] = true
			}
		}
	}
	for k, v := range want {
		if !v {
			t.Fatalf("expected pass for %s; stdout=\n%s", k, string(out))
		}
	}
}
