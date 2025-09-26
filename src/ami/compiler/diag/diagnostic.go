package diag

import srcset "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Diagnostic is the compiler-internal diagnostic format with optional position.
type Diagnostic struct {
    Level   Level
    Code    string
    Message string
    Package string
    File    string
    Pos     *srcset.Position
    EndPos  *srcset.Position
    Data    map[string]interface{}
}

