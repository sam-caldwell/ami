package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// Parser implements a minimal recursive-descent parser.
type Parser struct {
    s   *scanner.Scanner
    cur token.Token
    // collected leading comments to attach to the next node
    pending []ast.Comment
    errors  []error
    // decorators collected immediately before a function declaration
    pendingDecos []ast.Decorator
}

