package ir

import astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// Module is the IR container for a single source unit.
type Module struct {
    Package   string
    Version   string
    Unit      string // file path
    Functions []Function
    AST       *astpkg.File
    // Directive-derived attributes (scaffold)
    Concurrency  int
    Capabilities []string
    Trust        string
    Backpressure string
    Pipelines    []PipelineIR
}

