package workspace

import (
    "gopkg.in/yaml.v3"
    "testing"
    "strings"
)

func TestToolchain_YAMLKeysPresent(t *testing.T) {
    tc := Toolchain{
        Compiler: Compiler{Concurrency: "NUM_CPU", Target: "./build", Env: []string{"darwin/arm64"}, Options: []string{"verbose"}},
        Linker:   Linker{Options: []string{"Optimize: 0"}},
        Linter:   Linter{Options: []string{"strict"}, Rules: map[string]string{"W_IMPORT_ORDER": "warn"}},
    }
    b, err := yaml.Marshal(tc)
    if err != nil { t.Fatalf("yaml marshal: %v", err) }
    s := string(b)
    for _, k := range []string{"compiler:", "linker:", "linter:"} {
        if !strings.Contains(s, k) { t.Fatalf("expected key %q in YAML: %s", k, s) }
    }
}

