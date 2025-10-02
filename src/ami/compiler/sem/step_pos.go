package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

func stepPos(st *ast.StepStmt) diag.Position {
    return diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
}

