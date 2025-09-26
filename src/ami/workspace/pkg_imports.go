package workspace

import (
    "errors"
    "fmt"
    "strings"
)

// validatePackageImports inspects packages[*].<pkg>.import (if present) and
// validates entries as "<path> [constraint]" with supported operators.
func (w *Workspace) validatePackageImports() error {
    for _, p := range w.Packages {
        m, ok := p.(map[string]any)
        if !ok {
            continue
        }
        for _, v := range m {
            pm, ok := v.(map[string]any)
            if !ok {
                continue
            }
            imp, ok := pm["import"]
            if !ok || imp == nil {
                continue
            }
            lst, ok := imp.([]any)
            if !ok {
                return errors.New("packages.import must be a sequence of strings")
            }
            for _, item := range lst {
                s, ok := item.(string)
                if !ok {
                    return errors.New("packages.import items must be strings")
                }
                fields := strings.Fields(s)
                if len(fields) == 0 {
                    return errors.New("packages.import contains empty entry")
                }
                if len(fields) > 2 {
                    return fmt.Errorf("invalid import entry (too many tokens): %q", s)
                }
                // path is fields[0], constraint optional
                if len(fields) == 2 {
                    cons := strings.ReplaceAll(fields[1], " ", "")
                    if !isValidConstraint(cons) {
                        return fmt.Errorf("invalid version constraint: %q", fields[1])
                    }
                }
            }
        }
    }
    return nil
}

func isValidConstraint(s string) bool {
    if s == "" || s == "==latest" {
        return true
    }
    if semverRe.MatchString(s) {
        return true
    }
    if strings.HasPrefix(s, "^") {
        return semverRe.MatchString(strings.TrimPrefix(s, "^"))
    }
    if strings.HasPrefix(s, "~") {
        return semverRe.MatchString(strings.TrimPrefix(s, "~"))
    }
    if strings.HasPrefix(s, ">=") {
        return semverRe.MatchString(strings.TrimPrefix(s, ">="))
    }
    if strings.HasPrefix(s, ">") {
        return semverRe.MatchString(strings.TrimPrefix(s, ">"))
    }
    return false
}

