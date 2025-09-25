package ast

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

// Position is a lightweight source position.
type Position struct{ Line, Column, Offset int }

// Comment attaches source comments to nodes with a starting position.
type Comment struct {
	Text string
	Pos  Position
}

// File represents a parsed source file.
// Backward-compatible fields Package/Imports remain for existing tests.
type File struct {
	Package    string
	Version    string
	Imports    []string
	Decls      []Node
	Stmts      []Node // legacy; will mirror Decls for now
	Directives []Directive
}

type Node interface{ isNode() }

// Bad node for unparsed tokens
type Bad struct{ Tok tok.Token }

func (Bad) isNode() {}

// ImportDecl captures an import with optional alias.
type ImportDecl struct {
	Path       string
	Alias      string // optional
	Constraint string // optional version constraint (e.g., ">= v1.2.3")
	Pos        Position
	Comments   []Comment
}

func (ImportDecl) isNode() {}

// FuncDecl scaffold for function declarations.
type FuncDecl struct {
	Name   string
	Params []Param
	Result []TypeRef
	// Body is omitted in scaffold; parser skips over it
	Body      []tok.Token // captured tokens for simple semantic checks (e.g., mutability)
	BodyStmts []Stmt      // simple statement AST parsed from tokens (scaffold)
	Pos       Position
	Comments  []Comment
}

func (FuncDecl) isNode() {}

type Param struct {
	Name string
	Type TypeRef
}
type TypeRef struct {
	Name   string
	Args   []TypeRef // generic arguments, e.g., Event<T>
	Ptr    bool      // pointer
	Slice  bool      // slice []
	Offset int       // source offset where this type begins
}

// EnumDecl scaffold
type EnumDecl struct {
	Name     string
	Members  []EnumMember
	Pos      Position
	Comments []Comment
}
type EnumMember struct {
	Name  string
	Value string
}

func (EnumDecl) isNode() {}

// StructDecl scaffold
type StructDecl struct {
	Name     string
	Fields   []Field
	Pos      Position
	Comments []Comment
}
type Field struct {
	Name string
	Type TypeRef
}

func (StructDecl) isNode() {}

// PipelineDecl scaffold: e.g., Pipeline(name) { ... }
type PipelineDecl struct {
	Name            string
	Steps           []NodeCall
	Connectors      []string // between steps: "." or "->"
	ErrorSteps      []NodeCall
	ErrorConnectors []string
	Pos             Position
	Comments        []Comment
}

func (PipelineDecl) isNode() {}

// NodeCall represents a node invocation in a pipeline chain.
type NodeCall struct {
	Name     string
	Args     []string // raw argument expressions (scaffold)
	Workers  []WorkerRef
	Pos      Position
	Comments []Comment
}

// WorkerRef represents a referenced worker in a pipeline step argument list.
// Kind is "function" or "factory".
type WorkerRef struct {
	Name string
	Kind string
}

// Directive captures a top-level `#pragma` directive and payload.
type Directive struct {
	Name     string
	Payload  string
	Pos      Position
	Comments []Comment
}

func (Directive) isNode() {}

// --- Additional AST scaffolding for declarations, expressions, edges, and types ---

// PackageDecl represents a package declaration.
type PackageDecl struct{ Name string }

func (PackageDecl) isNode() {}

// Expressions
type Expr interface{ isExpr() }

type Ident struct {
	Name string
	Pos  Position
}

func (Ident) isExpr() {}

type BasicLit struct {
	Kind  string
	Value string
	Pos   Position
}

func (BasicLit) isExpr() {}

type CallExpr struct {
	Fun  Expr
	Args []Expr
	Pos  Position
}

func (CallExpr) isExpr() {}

// BinaryExpr represents a binary operation: X Op Y
type BinaryExpr struct {
	X   Expr
	Op  string
	Y   Expr
	Pos Position
}

func (BinaryExpr) isExpr() {}

// UnaryExpr represents a simple unary operation like &x or *p
type UnaryExpr struct {
	Op  string
	X   Expr
	Pos Position
}

func (UnaryExpr) isExpr() {}

// SelectorExpr represents a qualified selector: recv.method
type SelectorExpr struct {
	X   Expr
	Sel string
	Pos Position
}

func (SelectorExpr) isExpr() {}

// --- Simple statement nodes for function bodies ---

type Stmt interface{ isStmt() }

type ExprStmt struct {
	X        Expr
	Pos      Position
	Comments []Comment
}

func (ExprStmt) isStmt() {}

type AssignStmt struct {
	LHS      Expr
	RHS      Expr
	Pos      Position
	Comments []Comment
}

func (AssignStmt) isStmt() {}

// DeferStmt represents a deferred execution of an expression (usually a call).
type DeferStmt struct {
	X        Expr
	Pos      Position
	Comments []Comment
}

func (DeferStmt) isStmt() {}

type BlockStmt struct {
	Stmts    []Stmt
	Pos      Position
	Comments []Comment
}

func (BlockStmt) isStmt() {}

type MutBlockStmt struct {
	Body     BlockStmt
	Pos      Position
	Comments []Comment
}

func (MutBlockStmt) isStmt() {}

// ReturnStmt returns from a function with zero or more expressions (tuple future).
type ReturnStmt struct {
	Results  []Expr
	Pos      Position
	Comments []Comment
}

func (ReturnStmt) isStmt() {}

// VarDeclStmt declares a local variable with optional type and initializer.
type VarDeclStmt struct {
	Name     string
	Type     TypeRef // optional (zero value if omitted)
	Init     Expr    // optional (nil if omitted)
	Pos      Position
	Comments []Comment
}

func (VarDeclStmt) isStmt() {}

// ContainerLit represents container literals like:
// - slice<T>{ e1, e2, ... }
// - set<T>{ e1, e2, ... }
// - map<K,V>{ k1: v1, k2: v2, ... }
type ContainerLit struct {
	Kind     string    // "slice" | "set" | "map"
	TypeArgs []TypeRef // 1 for slice/set, 2 for map
	Elems    []Expr    // slice/set elements
	MapElems []MapElem // map elements
	Pos      Position
}

func (ContainerLit) isExpr() {}

type MapElem struct {
	Key   Expr
	Value Expr
}

// EdgeSpec captures an explicit edge with policy/config (future use).
type EdgeSpec struct {
	From         string
	To           string
	Connector    string // "." or "->"
	Backpressure string // e.g., drop/block; optional
	BufferSize   int    // optional
}

func (EdgeSpec) isNode() {}
