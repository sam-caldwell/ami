package root

import astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// lintUnitRec represents a discovered source unit
type lintUnitRec struct {
    pkg, file, src string
    imports        []string
    ast            *astpkg.File
}

// lintConfig represents runtime linter options sourced from the workspace
// and in-file pragmas.
type lintConfig struct {
    // severity maps rule code -> desired level (error|warn|info). Missing→use default.
    // The string "off" in the workspace disables the rule globally.
    severity map[string]string
    // strict preset from workspace config (in addition to --strict flag)
    strict bool
    // suppressions: package name -> set(rule)
    suppressPkg map[string]map[string]bool
    // path-based suppressions: list of {glob, rules}
    suppressPaths []struct {
        glob  string
        rules map[string]bool
    }
}

