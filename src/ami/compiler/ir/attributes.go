package ir

import (
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// ApplyDirectives sets module attributes derived from top-level pragmas.
func (m *Module) ApplyDirectives(dirs []astpkg.Directive) {
    for _, d := range dirs {
        switch strings.ToLower(d.Name) {
        case "capabilities":
            m.Capabilities = splitCSV(d.Payload)
        case "trust":
            m.Trust = strings.TrimSpace(d.Payload)
        case "backpressure":
            m.Backpressure = strings.TrimSpace(d.Payload)
        }
    }
}

func splitCSV(s string) []string {
    var out []string
    for _, p := range strings.Split(s, ",") {
        v := strings.TrimSpace(p)
        if v != "" {
            out = append(out, v)
        }
    }
    return out
}

