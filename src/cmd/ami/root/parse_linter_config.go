package root

import (
    "path/filepath"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func parseLinterConfig(ws *workspace.Workspace) lintConfig {
    cfg := lintConfig{severity: map[string]string{}, suppressPkg: map[string]map[string]bool{}}
    // workspace settings (if present)
    // Example structure: toolchain: { linter: { strict: true, severity: {"W_X": "error"}, suppress: { packages: {"main": ["W_X"]}, paths: [{glob: "**/*.ami", rules: ["W_Y"]}] } } }
    // Forgiving parsing; treat unknown shapes as absent
    if m, ok := ws.Toolchain.Linter.(map[string]any); ok {
        if b, ok := m["strict"].(bool); ok {
            cfg.strict = b
        }
        if sev, ok := m["severity"].(map[string]any); ok {
            for k, v := range sev {
                if s, ok2 := v.(string); ok2 {
                    cfg.severity[k] = strings.ToLower(strings.TrimSpace(s))
                }
            }
        }
        if sup, ok := m["suppress"].(map[string]any); ok {
            if pkgs, ok := sup["packages"].(map[string]any); ok {
                for name, val := range pkgs {
                    lst, _ := val.([]any)
                    set := map[string]bool{}
                    for _, it := range lst {
                        if s, ok := it.(string); ok {
                            set[s] = true
                        }
                    }
                    if len(set) > 0 {
                        cfg.suppressPkg[name] = set
                    }
                }
            }
            if paths, ok := sup["paths"].([]any); ok {
                for _, it := range paths {
                    if m2, ok := it.(map[string]any); ok {
                        glob, _ := m2["glob"].(string)
                        raw, _ := m2["rules"].([]any)
                        rules := map[string]bool{}
                        for _, r := range raw {
                            if s, ok := r.(string); ok {
                                rules[s] = true
                            }
                        }
                        if strings.TrimSpace(glob) != "" && len(rules) > 0 {
                            // normalize glob to OS-specific path patterns if needed
                            cfg.suppressPaths = append(cfg.suppressPaths, struct {
                                glob  string
                                rules map[string]bool
                            }{glob: filepath.Clean(glob), rules: rules})
                        }
                    }
                }
            }
        }
    }
    return cfg
}

