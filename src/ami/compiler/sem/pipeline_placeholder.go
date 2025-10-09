package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// isNoopPlaceholder returns true when a step is written as `Name()` with
// no arguments and no attributes, signaling a placeholder pass-through stage.
func isNoopPlaceholder(st *ast.StepStmt) bool {
    if st == nil { return false }
    // Parens present and empty; no attributes
    hasParens := st.LParen.Line != 0 && st.RParen.Line != 0
    return hasParens && len(st.Args) == 0 && len(st.Attrs) == 0
}

