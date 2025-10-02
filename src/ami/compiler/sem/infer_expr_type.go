package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// inferExprType is an adapter used by other analyzers that need env-aware inference.
func inferExprType(env map[string]string, e ast.Expr) string { return inferLocalExprType(env, e) }

