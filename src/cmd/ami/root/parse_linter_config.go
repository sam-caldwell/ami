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
    // Forgiving parsing; support legacy/simplified shapes used in tests:
    // - suppress.package: { name: [rules] }
    // - suppress.paths: { "glob": [rules] }  OR  [ {glob, rules} ]
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
        // Backward/alias support: allow "rules" to specify severities
        if sev2, ok := m["rules"].(map[string]any); ok {
            for k, v := range sev2 {
                if s, ok2 := v.(string); ok2 {
                    cfg.severity[k] = strings.ToLower(strings.TrimSpace(s))
                }
            }
        }
        if sup, ok := m["suppress"].(map[string]any); ok {
            // package or packages map
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
            // accept singular key "package" as alias for backward-compat
            if pkgs1, ok := sup["package"].(map[string]any); ok {
                for name, val := range pkgs1 {
                    lst, _ := val.([]any)
                    set := map[string]bool{}
                    for _, it := range lst {
                        if s, ok := it.(string); ok {
                            set[s] = true
                        }
                    }
                    if len(set) > 0 {
                        if cfg.suppressPkg[name] == nil {
                            cfg.suppressPkg[name] = map[string]bool{}
                        }
                        for k := range set {
                            cfg.suppressPkg[name][k] = true
                        }
                    }
                }
            }
            // paths: accept list of objects [{glob, rules}] (new) OR map[glob][]rules (legacy)
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
                            cfg.suppressPaths = append(cfg.suppressPaths, struct {
                                glob  string
                                rules map[string]bool
                            }{glob: filepath.Clean(glob), rules: rules})
                        }
                    }
                }
            } else if pmap, ok := sup["paths"].(map[string]any); ok {
                for glob, v := range pmap {
                    raw, _ := v.([]any)
                    rules := map[string]bool{}
                    for _, r := range raw {
                        if s, ok := r.(string); ok {
                            rules[s] = true
                        }
                    }
                    if strings.TrimSpace(glob) != "" && len(rules) > 0 {
                        cfg.suppressPaths = append(cfg.suppressPaths, struct {
                            glob  string
                            rules map[string]bool
                        }{glob: filepath.Clean(glob), rules: rules})
                    }
                }
            }
        }
    }
    return cfg
}
