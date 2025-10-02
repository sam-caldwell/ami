package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestGraphFromPipeline_FilePair(t *testing.T) {
    pd := &ast.PipelineDecl{}
    _ = graphFromPipeline("pkg", "unit", pd)
}

