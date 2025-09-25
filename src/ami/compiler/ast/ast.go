package ast

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

// Position is a lightweight source position.
type Position struct { Line, Column, Offset int }

// File represents a parsed source file.
// Backward-compatible fields Package/Imports remain for existing tests.
type File struct {
    Package string
    Imports []string
    Decls   []Node
    Stmts   []Node // legacy; will mirror Decls for now
    Directives []Directive
}

type Node interface{ isNode() }

// Bad node for unparsed tokens
type Bad struct{ Tok tok.Token }
func (Bad) isNode() {}

// ImportDecl captures an import with optional alias.
type ImportDecl struct {
    Path  string
    Alias string // optional
}
func (ImportDecl) isNode() {}

// FuncDecl scaffold for function declarations.
type FuncDecl struct {
    Name   string
    Params []Param
    Result []TypeRef
    // Body is omitted in scaffold; parser skips over it
}
func (FuncDecl) isNode() {}

type Param struct { Name string; Type TypeRef }
type TypeRef struct { Name string }

// EnumDecl scaffold
type EnumDecl struct { Name string; Members []EnumMember }
type EnumMember struct { Name string; Value string }
func (EnumDecl) isNode() {}

// StructDecl scaffold
type StructDecl struct { Name string; Fields []Field }
type Field struct { Name string; Type TypeRef }
func (StructDecl) isNode() {}

// PipelineDecl scaffold: e.g., Pipeline(name) { ... }
type PipelineDecl struct {
    Name       string
    Steps      []NodeCall
    Connectors []string // between steps: "." or "->"
    ErrorSteps []NodeCall
    ErrorConnectors []string
}
func (PipelineDecl) isNode() {}

// NodeCall represents a node invocation in a pipeline chain.
type NodeCall struct {
    Name string
    Args []string // raw argument expressions (scaffold)
}

// Directive captures a top-level `#pragma` directive and payload.
type Directive struct {
    Name    string
    Payload string
}
func (Directive) isNode() {}

// --- Additional AST scaffolding for declarations, expressions, edges, and types ---

// PackageDecl represents a package declaration.
type PackageDecl struct { Name string }
func (PackageDecl) isNode() {}

// Expressions
type Expr interface{ isExpr() }

type Ident struct { Name string }
func (Ident) isExpr() {}

type BasicLit struct { Kind string; Value string }
func (BasicLit) isExpr() {}

type CallExpr struct { Fun Expr; Args []Expr }
func (CallExpr) isExpr() {}

// EdgeSpec captures an explicit edge with policy/config (future use).
type EdgeSpec struct {
    From       string
    To         string
    Connector  string // "." or "->"
    Backpressure string // e.g., drop/block; optional
    BufferSize int      // optional
}
func (EdgeSpec) isNode() {}
