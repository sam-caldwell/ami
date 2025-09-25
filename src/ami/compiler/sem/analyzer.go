package sem

import (
	"fmt"
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/diag"
	srcset "github.com/sam-caldwell/ami/src/ami/compiler/source"
	tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
	"github.com/sam-caldwell/ami/src/ami/compiler/types"
	"strconv"
	"strings"
)

type Result struct {
	Scope       *types.Scope
	Diagnostics []diag.Diagnostic
}

// AnalyzeFile performs minimal semantic analysis:
// - Build a top-level scope
// - Detect duplicate function declarations
// - Validate basic pipeline semantics (ingress→...→egress)
func AnalyzeFile(f *astpkg.File) Result {
	res := Result{Scope: types.NewScope(nil)}
	seen := map[string]bool{}
	// collect function names for worker resolution
	funcs := map[string]astpkg.FuncDecl{}
	// Insert imported package symbols for name resolution (alias or last path segment)
	for _, d := range f.Decls {
		if id, ok := d.(astpkg.ImportDecl); ok {
			name := id.Alias
			if name == "" {
				// derive from path last segment
				parts := strings.Split(id.Path, "/")
				name = parts[len(parts)-1]
			}
			// placeholder type for imported package symbols
			_ = res.Scope.Insert(&types.Object{Kind: types.ObjType, Name: name, Type: types.TPackage})
		}
	}
	for _, d := range f.Decls {
		if fd, ok := d.(astpkg.FuncDecl); ok {
			if fd.Name == "_" {
				res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IDENT_ILLEGAL", Message: "blank identifier '_' cannot be used as a function name"})
				continue
			}
			// parameters: disallow blank identifier as a parameter name
			for _, p := range fd.Params {
				if p.Name == "_" {
					res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_PARAM_ILLEGAL", Message: "blank identifier '_' cannot be used as a parameter name"})
				}
			}
			if seen[fd.Name] {
				res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{
					Level:   diag.Error,
					Code:    "E_DUP_FUNC",
					Message: fmt.Sprintf("duplicate function %q", fd.Name),
					File:    "",
				})
				continue
			}
			// Build a function type from parameter/result TypeRefs
			var ps []types.Type
			for _, p := range fd.Params {
				ps = append(ps, types.FromAST(p.Type))
			}
			var rs []types.Type
			for _, r := range fd.Result {
				rs = append(rs, types.FromAST(r))
			}
			_ = res.Scope.Insert(&types.Object{Kind: types.ObjFunc, Name: fd.Name, Type: types.Function{Params: ps, Results: rs}})
			seen[fd.Name] = true
			funcs[fd.Name] = fd
			// Mutability: enforce AMI semantics (no mut blocks; '*' LHS marker required)
			res.Diagnostics = append(res.Diagnostics, analyzeMutationMarkers(fd)...)
			// Pointer/address prohibitions (2.3.2): '&' not allowed
			res.Diagnostics = append(res.Diagnostics, analyzePointerProhibitions(fd)...)
			// Imperative type checks (2.3): simple assignment type rules (no raw pointers)
			res.Diagnostics = append(res.Diagnostics, analyzeImperativeTypes(fd)...)
			// Call argument type checks with simple generic unification
			res.Diagnostics = append(res.Diagnostics, analyzeCallTypes(fd, funcs)...)
			// Operators: arithmetic/comparison operand compatibility
			res.Diagnostics = append(res.Diagnostics, analyzeOperators(fd)...)
			// Event contracts (1.7): event parameter immutability
			res.Diagnostics = append(res.Diagnostics, analyzeEventContracts(fd)...)
			// State contracts (2.2.14/2.3.5): state param immutability/address-of
			res.Diagnostics = append(res.Diagnostics, analyzeStateContracts(fd)...)
			// Memory domains (2.4): forbid cross-domain references into state
			res.Diagnostics = append(res.Diagnostics, analyzeMemoryDomains(fd)...)
			// Memory model (2.4): ownership & RAII scaffolding
			res.Diagnostics = append(res.Diagnostics, analyzeRAII(fd, funcs)...)
		}
		if ed, ok := d.(astpkg.EnumDecl); ok {
			res.Diagnostics = append(res.Diagnostics, analyzeEnum(ed)...)
		}
		if sd, ok := d.(astpkg.StructDecl); ok {
			res.Diagnostics = append(res.Diagnostics, analyzeStruct(sd)...)
		}
		if id, ok := d.(astpkg.ImportDecl); ok {
			if id.Alias == "_" {
				res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IMPORT_ALIAS", Message: "blank identifier '_' cannot be used as import alias"})
			}
		}
		if pd, ok := d.(astpkg.PipelineDecl); ok {
			res.Diagnostics = append(res.Diagnostics, analyzePipeline(pd)...)
			res.Diagnostics = append(res.Diagnostics, analyzeWorkers(pd, funcs, res.Scope)...)
			res.Diagnostics = append(res.Diagnostics, analyzeEventTypeFlow(pd, funcs)...)
			res.Diagnostics = append(res.Diagnostics, analyzeIOPermissions(pd)...)
			res.Diagnostics = append(res.Diagnostics, analyzeEdges(pd)...)
			res.Diagnostics = append(res.Diagnostics, analyzeEdgeTypeSafety(pd, funcs)...)
		}
	}
	// Global type checks
	res.Diagnostics = append(res.Diagnostics, analyzeMapTypes(f)...)
	res.Diagnostics = append(res.Diagnostics, analyzeSetTypes(f)...)
	res.Diagnostics = append(res.Diagnostics, analyzeSliceTypes(f)...)
	// Memory domains (2.4 scaffold) across functions
	for _, d := range f.Decls {
		if fd, ok := d.(astpkg.FuncDecl); ok {
			res.Diagnostics = append(res.Diagnostics, analyzeMemoryDomains(fd)...)
		}
	}
	// Cross-pipeline cycle detection (unless cycle pragma present)
	res.Diagnostics = append(res.Diagnostics, analyzeCycles(f)...)
	return res
}

// analyzeEnum validates enum declarations: non-empty members, unique names,
// valid literal values (if provided), and disallow blank identifier members.
func analyzeEnum(ed astpkg.EnumDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if ed.Name == "" {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_NAME", Message: "enum must have a name"})
	}
	if len(ed.Members) == 0 {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_EMPTY", Message: "enum has no members"})
		return diags
	}
	seen := map[string]bool{}
	for _, m := range ed.Members {
		if m.Name == "_" {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_BLANK_MEMBER", Message: "enum member cannot be '_'"})
		}
		if seen[m.Name] {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_DUP_MEMBER", Message: "duplicate enum member: " + m.Name})
		}
		seen[m.Name] = true
		if m.Value != "" {
			if !(isIntLiteral(m.Value) || isStringLiteral(m.Value)) {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_VALUE_INVALID", Message: "enum member value must be integer or string literal: " + m.Name})
			}
		}
	}
	return diags
}

func isIntLiteral(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	if s[0] == '-' {
		if len(s) == 1 {
			return false
		}
		i = 1
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
func isStringLiteral(s string) bool {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return true
	}
	return false
}

// analyzeStruct validates struct declarations: non-empty fields, unique names,
// non-blank field names, and presence of a type on each field.
func analyzeStruct(sd astpkg.StructDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if sd.Name == "" {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_NAME", Message: "struct must have a name"})
	}
	if len(sd.Fields) == 0 {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_EMPTY", Message: "struct has no fields"})
		return diags
	}
	seen := map[string]bool{}
	for _, f := range sd.Fields {
		if f.Name == "_" {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_BLANK_FIELD", Message: "struct field cannot be '_'"})
		}
		if f.Name == "" {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_FIELD_NAME", Message: "struct field must have a name"})
		}
		if seen[f.Name] {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_DUP_FIELD", Message: "duplicate struct field: " + f.Name})
		}
		seen[f.Name] = true
		if f.Type.Name == "" && !f.Type.Ptr && !f.Type.Slice { // no recognizable type
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_FIELD_TYPE_INVALID", Message: "struct field missing or invalid type: " + f.Name})
		}
	}
	return diags
}

// analyzeMapTypes walks declared function signatures and struct fields to ensure
// any `map<K,V>` types meet minimal constraints: exactly two type arguments, and
// key type K is not a pointer, slice, map, set, or slice, and has no generic args.
func analyzeMapTypes(f *astpkg.File) []diag.Diagnostic {
	var diags []diag.Diagnostic
	var walk func(t astpkg.TypeRef)
	walk = func(t astpkg.TypeRef) {
		if strings.ToLower(t.Name) == "map" {
			if len(t.Args) != 2 {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_ARITY", Message: "map must have exactly two type arguments: map<K,V>"})
			} else {
				k := t.Args[0]
				if k.Ptr || k.Slice {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be pointer or slice"})
				}
				switch strings.ToLower(k.Name) {
				case "map", "set", "slice":
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be map/set/slice"})
				}
				if len(k.Args) > 0 {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be generic"})
				}
			}
		}
		for _, a := range t.Args {
			walk(a)
		}
	}
	for _, d := range f.Decls {
		if sd, ok := d.(astpkg.StructDecl); ok {
			for _, fld := range sd.Fields {
				walk(fld.Type)
			}
		}
		if fd, ok := d.(astpkg.FuncDecl); ok {
			for _, p := range fd.Params {
				walk(p.Type)
			}
			for _, r := range fd.Result {
				walk(r)
			}
		}
	}
	return diags
}

// analyzeSetTypes walks declared function signatures and struct fields to ensure
// any `set<T>` types meet minimal constraints: exactly one type argument, and
// element type T is not a pointer, slice, map, set, or slice, and has no generic args.
func analyzeSetTypes(f *astpkg.File) []diag.Diagnostic {
	var diags []diag.Diagnostic
	var walk func(t astpkg.TypeRef)
	walk = func(t astpkg.TypeRef) {
		if strings.ToLower(t.Name) == "set" {
			if len(t.Args) != 1 {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SET_ARITY", Message: "set must have exactly one type argument: set<T>"})
			} else {
				e := t.Args[0]
				if e.Ptr || e.Slice {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SET_ELEM_TYPE_INVALID", Message: "set element type cannot be pointer or slice"})
				}
				switch strings.ToLower(e.Name) {
				case "map", "set", "slice":
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SET_ELEM_TYPE_INVALID", Message: "set element type cannot be map/set/slice"})
				}
				if len(e.Args) > 0 {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SET_ELEM_TYPE_INVALID", Message: "set element type cannot be generic"})
				}
			}
		}
		for _, a := range t.Args {
			walk(a)
		}
	}
	for _, d := range f.Decls {
		if sd, ok := d.(astpkg.StructDecl); ok {
			for _, fld := range sd.Fields {
				walk(fld.Type)
			}
		}
		if fd, ok := d.(astpkg.FuncDecl); ok {
			for _, p := range fd.Params {
				walk(p.Type)
			}
			for _, r := range fd.Result {
				walk(r)
			}
		}
	}
	return diags
}

// analyzeSliceTypes validates generic slice forms `slice<T>` for correct arity.
// Bracket slices `[]T` are represented by TypeRef{Slice:true, Name:T} and do not
// require additional constraints beyond nested type validation (e.g., maps).
func analyzeSliceTypes(f *astpkg.File) []diag.Diagnostic {
	var diags []diag.Diagnostic
	var walk func(t astpkg.TypeRef)
	walk = func(t astpkg.TypeRef) {
		if strings.ToLower(t.Name) == "slice" {
			if len(t.Args) != 1 {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SLICE_ARITY", Message: "slice must have exactly one type argument: slice<T>"})
			}
		}
		for _, a := range t.Args {
			walk(a)
		}
	}
	for _, d := range f.Decls {
		if sd, ok := d.(astpkg.StructDecl); ok {
			for _, fld := range sd.Fields {
				walk(fld.Type)
			}
		}
		if fd, ok := d.(astpkg.FuncDecl); ok {
			for _, p := range fd.Params {
				walk(p.Type)
			}
			for _, r := range fd.Result {
				walk(r)
			}
		}
	}
	return diags
}

func analyzePipeline(pd astpkg.PipelineDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	// Basic pipeline shape checks
	if len(pd.Steps) == 0 {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_EMPTY", Message: fmt.Sprintf("pipeline %q has no steps", pd.Name)})
		return diags
	}
	allowed := map[string]bool{"ingress": true, "transform": true, "fanout": true, "collect": true, "egress": true}
	ingressCount := 0
	egressCount := 0
	for i, step := range pd.Steps {
		name := strings.ToLower(step.Name)
		if !allowed[name] {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: fmt.Sprintf("unknown node %q", step.Name)})
			continue
		}
		switch name {
		case "ingress":
			ingressCount++
			if i != 0 {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_INGRESS_POSITION", Message: fmt.Sprintf("pipeline %q: ingress must be first", pd.Name)})
			}
		case "egress":
			egressCount++
			if i != len(pd.Steps)-1 {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EGRESS_POSITION", Message: fmt.Sprintf("pipeline %q: egress must be last", pd.Name)})
			}
		}
	}
	if ingressCount == 0 {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_START_INGRESS", Message: fmt.Sprintf("pipeline %q must start with ingress", pd.Name)})
	}
	if egressCount == 0 {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_END_EGRESS", Message: fmt.Sprintf("pipeline %q must end with egress", pd.Name)})
	}
	if ingressCount > 1 {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_INGRESS", Message: fmt.Sprintf("pipeline %q has multiple ingress nodes", pd.Name)})
	}
	if egressCount > 1 {
		diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_EGRESS", Message: fmt.Sprintf("pipeline %q has multiple egress nodes", pd.Name)})
	}

	// Error pipeline semantics
	if len(pd.ErrorSteps) > 0 {
		if strings.ToLower(pd.ErrorSteps[0].Name) == "ingress" {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ERRPIPE_START_INVALID", Message: fmt.Sprintf("pipeline %q error path cannot start with ingress", pd.Name)})
		}
		if strings.ToLower(pd.ErrorSteps[len(pd.ErrorSteps)-1].Name) != "egress" {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ERRPIPE_END_EGRESS", Message: fmt.Sprintf("pipeline %q error path must end with egress", pd.Name)})
		}
		for _, st := range pd.ErrorSteps {
			nm := strings.ToLower(st.Name)
			if !allowed[nm] {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: fmt.Sprintf("unknown node %q in error path", st.Name)})
			}
		}
	}
	return diags
}

// analyzeWorkers ensures worker/factory references in pipeline steps resolve
// to known top-level function declarations. It scans step args heuristically:
// - IDENT or IDENT(arg,...) → resolves to IDENT
// Applies to Transform and FanOut nodes.
func analyzeWorkers(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl, scope *types.Scope) []diag.Diagnostic {
	var diags []diag.Diagnostic
	checkArgs := func(args []string) {
		for _, a := range args {
			name := a
			hasCall := false
			if i := strings.IndexRune(a, '('); i >= 0 {
				name = strings.TrimSpace(a[:i])
				hasCall = true
			}
			// simple identifier extract: letters/_ followed by letters/digits/_
			if name == "" {
				continue
			}
			// skip placeholders like "cfg" or literals
			if name == "cfg" {
				continue
			}
			// Only enforce for explicit calls (factory invocations) or names starting with New
			if !(hasCall || strings.HasPrefix(name, "New")) {
				// if bare name, only check signature if function exists; otherwise skip
				if fd, ok := funcs[name]; ok {
					if !isWorkerSignature(fd) {
						diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: fmt.Sprintf("worker %q has invalid signature", name)})
					}
				}
				continue
			}
			// Dotted reference: pkg.Func(...) — accept if pkg is an imported package symbol
			if dot := strings.IndexByte(name, '.'); dot > 0 {
				pkg := name[:dot]
				if scope != nil {
					if obj := scope.Lookup(pkg); obj != nil && obj.Type.String() == types.TPackage.String() {
						// Imported worker reference; cannot validate signature here; skip undefined error.
						continue
					}
				}
			}
			fd, ok := funcs[name]
			if !ok {
				// allow blank identifier '_' to pass worker ref check as sink
				if name == "_" {
					continue
				}
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_UNDEFINED", Message: fmt.Sprintf("unknown worker/factory %q", name)})
			} else {
				if !isWorkerSignature(fd) {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: fmt.Sprintf("worker %q has invalid signature", name)})
				}
			}
		}
	}
	for _, st := range pd.Steps {
		n := strings.ToLower(st.Name)
		switch n {
		case "transform", "fanout":
			checkArgs(st.Args)
		}
	}
	for _, st := range pd.ErrorSteps {
		n := strings.ToLower(st.Name)
		switch n {
		case "transform", "fanout":
			checkArgs(st.Args)
		}
	}
	return diags
}

// analyzeMutationMarkers enforces AMI mutability rules:
// - No Rust-like `mut { ... }` blocks are permitted.
// - Any assignment must use `*` on the left-hand side to mark mutation.
// - Unary '*' is not a dereference and is invalid in expression (RHS) position.
// Note: In AMI 2.3.2, `*` on the LHS is the mutation marker (not a pointer
// dereference). This function validates that usage and flags misuse.
func analyzeMutationMarkers(fd astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if len(fd.Body) == 0 && len(fd.BodyStmts) == 0 {
		return diags
	}
	if len(fd.BodyStmts) > 0 {
		var walkExpr func(astpkg.Expr, bool)
		walkExpr = func(e astpkg.Expr, isLHS bool) {
			switch v := e.(type) {
			case astpkg.UnaryExpr:
				if v.Op == "*" {
					if !isLHS {
						diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STAR_MISUSED", Message: "'*' is not a dereference; only allowed on assignment left-hand side as a mutability marker"})
					}
					walkExpr(v.X, isLHS)
				}
				// '&' is handled by parser; ignore here
			case astpkg.CallExpr:
				for _, a := range v.Args {
					walkExpr(a, false)
				}
			case astpkg.SelectorExpr:
				walkExpr(v.X, false)
			}
		}
		var walkStmt func(astpkg.Stmt)
		walkStmt = func(s astpkg.Stmt) {
			switch v := s.(type) {
			case astpkg.AssignStmt:
				if ue, ok := v.LHS.(astpkg.UnaryExpr); !ok || ue.Op != "*" {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_ASSIGN_UNMARKED", Message: "assignment must be marked with '*' (mutation marker) on left-hand side"})
				}
				walkExpr(v.RHS, false)
			case astpkg.MutBlockStmt:
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_BLOCK_UNSUPPORTED", Message: "mut { ... } blocks are not part of AMI; use mutate(expr) or '*' on assignment LHS"})
				for _, ss := range v.Body.Stmts {
					walkStmt(ss)
				}
			case astpkg.BlockStmt:
				for _, ss := range v.Stmts {
					walkStmt(ss)
				}
			case astpkg.ExprStmt:
				walkExpr(v.X, false)
			}
		}
		for _, s := range fd.BodyStmts {
			walkStmt(s)
		}
		return diags
	}
	// Token-based fallback
	for _, t := range fd.Body {
		if t.Kind == tok.KW_MUT {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_BLOCK_UNSUPPORTED", Message: "mut { ... } blocks are not part of AMI; use mutate(expr) or '*' on assignment LHS"})
		}
	}
	toks := fd.Body
	for i := 0; i < len(toks); i++ {
		if toks[i].Kind == tok.ASSIGN {
			if !(i-2 >= 0 && toks[i-2].Kind == tok.STAR && toks[i-1].Kind == tok.IDENT) {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_ASSIGN_UNMARKED", Message: "assignment must be marked with '*' (mutation marker) on left-hand side"})
			}
		}
	}
	return diags
}

// analyzeEdges validates declarative edge specs provided via `in=edge.*(...)` args.
// Emits diagnostics for invalid capacity ordering, negative capacities, and unknown
// backpressure policies. For edge.Pipeline, also requires a non-empty name.
func analyzeEdges(pd astpkg.PipelineDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	checkArgs := func(args []string) {
		if spec, ok := parseEdgeSpecFromArgs(args); ok {
			switch v := spec.(type) {
			case fifoSpec:
				if v.Min < 0 {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_MINCAP", Message: "edge FIFO: minCapacity must be >= 0"})
				}
				if v.Max < v.Min {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_CAP_ORDER", Message: "edge FIFO: maxCapacity must be >= minCapacity"})
				}
				if v.BP != "" && v.BP != "block" && v.BP != "drop" {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_BP_INVALID", Message: "edge FIFO: invalid backpressure policy"})
				}
			case lifoSpec:
				if v.Min < 0 {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_MINCAP", Message: "edge LIFO: minCapacity must be >= 0"})
				}
				if v.Max < v.Min {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_CAP_ORDER", Message: "edge LIFO: maxCapacity must be >= minCapacity"})
				}
				if v.BP != "" && v.BP != "block" && v.BP != "drop" {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_BP_INVALID", Message: "edge LIFO: invalid backpressure policy"})
				}
			case pipeSpec:
				if v.Name == "" {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_NAME_REQUIRED", Message: "edge Pipeline: upstream name required"})
				}
				if v.Min < 0 {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_MINCAP", Message: "edge Pipeline: minCapacity must be >= 0"})
				}
				if v.Max < v.Min {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_CAP_ORDER", Message: "edge Pipeline: maxCapacity must be >= minCapacity"})
				}
				if v.BP != "" && v.BP != "block" && v.BP != "drop" {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_BP_INVALID", Message: "edge Pipeline: invalid backpressure policy"})
				}
			}
		}
	}
	for _, st := range pd.Steps {
		checkArgs(st.Args)
	}
	for _, st := range pd.ErrorSteps {
		checkArgs(st.Args)
	}
	return diags
}

// Minimal local spec structs to avoid cross-package dependency
type fifoSpec struct {
	Min, Max int
	BP, Type string
}
type lifoSpec struct {
	Min, Max int
	BP, Type string
}
type pipeSpec struct {
	Name     string
	Min, Max int
	BP, Type string
}

// parseEdgeSpecFromArgs: copy of tolerant parser used in IR lowering (simplified)
func parseEdgeSpecFromArgs(args []string) (interface{}, bool) {
	for _, a := range args {
		s := strings.TrimSpace(a)
		if !strings.HasPrefix(s, "in=") {
			continue
		}
		v := strings.TrimPrefix(s, "in=")
		if strings.HasPrefix(v, "edge.FIFO(") && strings.HasSuffix(v, ")") {
			params := parseKVList(v[len("edge.FIFO(") : len(v)-1])
			var f fifoSpec
			for k, val := range params {
				switch k {
				case "minCapacity":
					f.Min = atoiSafe(val)
				case "maxCapacity":
					f.Max = atoiSafe(val)
				case "backpressure":
					f.BP = val
				case "type":
					f.Type = val
				}
			}
			return f, true
		}
		if strings.HasPrefix(v, "edge.LIFO(") && strings.HasSuffix(v, ")") {
			params := parseKVList(v[len("edge.LIFO(") : len(v)-1])
			var l lifoSpec
			for k, val := range params {
				switch k {
				case "minCapacity":
					l.Min = atoiSafe(val)
				case "maxCapacity":
					l.Max = atoiSafe(val)
				case "backpressure":
					l.BP = val
				case "type":
					l.Type = val
				}
			}
			return l, true
		}
		if strings.HasPrefix(v, "edge.Pipeline(") && strings.HasSuffix(v, ")") {
			params := parseKVList(v[len("edge.Pipeline(") : len(v)-1])
			var p pipeSpec
			for k, val := range params {
				switch k {
				case "name":
					p.Name = val
				case "minCapacity":
					p.Min = atoiSafe(val)
				case "maxCapacity":
					p.Max = atoiSafe(val)
				case "backpressure":
					p.BP = val
				case "type":
					p.Type = val
				}
			}
			return p, true
		}
	}
	return nil, false
}

func parseKVList(s string) map[string]string {
	out := map[string]string{}
	parts := splitTopLevelCommas(s)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if eq := strings.IndexByte(p, '='); eq >= 0 {
			k := strings.TrimSpace(p[:eq])
			v := strings.TrimSpace(p[eq+1:])
			if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
				v = v[1 : len(v)-1]
			}
			out[k] = v
		}
	}
	return out
}

func splitTopLevelCommas(s string) []string {
	var out []string
	depth := 0
	last := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				out = append(out, s[last:i])
				last = i + 1
			}
		}
	}
	out = append(out, s[last:])
	return out
}

func atoiSafe(s string) int { n, _ := strconv.Atoi(s); return n }

// typeRefToString renders a TypeRef to a string including pointer, slice, and generics.
func typeRefToString(t astpkg.TypeRef) string {
	var b strings.Builder
	if t.Ptr {
		b.WriteByte('*')
	}
	if t.Slice {
		b.WriteString("[]")
	}
	b.WriteString(t.Name)
	if len(t.Args) > 0 {
		b.WriteByte('<')
		for i, a := range t.Args {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(typeRefToString(a))
		}
		b.WriteByte('>')
	}
	return b.String()
}

// analyzePointerProhibitions enforces AMI 2.3.2: no raw pointer/address operators.
// Specifically, disallow '&' anywhere in function bodies.
func analyzePointerProhibitions(fd astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if len(fd.Body) == 0 {
		return diags
	}
	for _, t := range fd.Body {
		if t.Kind == tok.AMP {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "'&' address-of operator is not allowed; AMI does not expose raw pointers (see 2.3.2)"})
		}
	}
	return diags
}

// analyzeEventContracts enforces event parameter immutability: the Event<T>
// parameter (commonly named 'ev') cannot be assigned.
func analyzeEventContracts(fd astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if len(fd.Params) < 2 || len(fd.Body) == 0 {
		return diags
	}
	// detect event parameter name
	evName := ""
	if p := fd.Params[1]; p.Type.Name == "Event" && len(p.Type.Args) == 1 && p.Name != "" {
		evName = p.Name
	}
	if evName == "" {
		return diags
	}
	toks := fd.Body
	for i := 0; i < len(toks); i++ {
		if toks[i].Kind == tok.ASSIGN {
			// LHS ident equal to event param, not a deref
			if i-1 >= 0 && toks[i-1].Kind == tok.IDENT && toks[i-1].Lexeme == evName {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EVENT_PARAM_ASSIGN", Message: "event parameter is immutable and cannot be reassigned"})
			}
		}
	}
	return diags
}

// analyzeStateContracts enforces basic node-state parameter rules:
// - State parameter is immutable (cannot be reassigned): E_STATE_PARAM_ASSIGN
// (No address-of in AMI 2.3.2)
func analyzeStateContracts(fd astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if len(fd.Params) < 3 || len(fd.Body) == 0 {
		return diags
	}
	// Build simple env of param name -> type
	env := map[string]astpkg.TypeRef{}
	for _, p := range fd.Params {
		if p.Name != "" {
			env[p.Name] = p.Type
		}
	}
	toks := fd.Body
	for i := 0; i < len(toks); i++ {
		// Reassignment of state param: IDENT '=' ... where IDENT is State
		if toks[i].Kind == tok.IDENT {
			if tr, ok := env[toks[i].Lexeme]; ok && tr.Name == "State" {
				if i+1 < len(toks) && toks[i+1].Kind == tok.ASSIGN {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STATE_PARAM_ASSIGN", Message: "state parameter is immutable and cannot be reassigned"})
				}
			}
		}
		// No address-of in AMI 2.3.2
	}
	return diags
}

// analyzeMemoryDomains enforces basic allocation domain separation (6.5/Ch.2.4):
// - Event heap (Event<T>), Node-state (State), Ephemeral stack (locals/others).
// Forbidden cross-domain references:
//   - Assigning address of non-state value into state memory, e.g., `*st = &ev` or `*st = &x`.
//     Emits E_CROSS_DOMAIN_REF.
func analyzeMemoryDomains(fd astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if len(fd.Body) == 0 {
		return diags
	}
	// Token-based scan for prohibited cross-domain patterns. We prefer token
	// matching to avoid coupling to expression forms and to work even when the
	// parser recorded address-of errors.
	env := map[string]astpkg.TypeRef{}
	for _, p := range fd.Params {
		if p.Name != "" {
			env[p.Name] = p.Type
		}
	}
	toks := fd.Body
	for i := 0; i+3 < len(toks); i++ {
		// *st = &x (address-of into state)
		if toks[i].Kind == tok.STAR && toks[i+1].Kind == tok.IDENT && toks[i+2].Kind == tok.ASSIGN && toks[i+3].Kind == tok.AMP {
			if tr, ok := env[toks[i+1].Lexeme]; ok && tr.Name == "State" {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CROSS_DOMAIN_REF", Message: "cross-domain reference into state: cannot assign address-of non-state value into state"})
			}
		}
		// *st = ev (assignment from non-state identifier into state)
		if toks[i].Kind == tok.STAR && toks[i+1].Kind == tok.IDENT && toks[i+2].Kind == tok.ASSIGN && toks[i+3].Kind == tok.IDENT {
			if tr, ok := env[toks[i+1].Lexeme]; ok && tr.Name == "State" {
				if rt, ok2 := env[toks[i+3].Lexeme]; ok2 && rt.Name != "State" {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CROSS_DOMAIN_REF", Message: "cross-domain assignment into state from non-state value is forbidden"})
				}
			}
		}
	}
	return diags
}

// analyzeRAII enforces minimal ownership/RAII rules for Owned<T> parameters:
// - E_RAII_OWNED_NOT_RELEASED: Owned param must be released or transferred.
// - E_RAII_DOUBLE_RELEASE: multiple releases/transfers for same variable.
// - E_RAII_USE_AFTER_RELEASE: use of variable after release/transfer.
// Release/transfer detection:
// - Calls to functions whose corresponding param type is Owned<…>.
// - Calls to known releasers: release(x), drop(x), free(x), dispose(x) or x.Close()/x.Release()/x.Free()/x.Dispose().
func analyzeRAII(fd astpkg.FuncDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if len(fd.Body) == 0 {
		return diags
	}
	if len(fd.BodyStmts) > 0 {
		return analyzeRAIIFromAST(fd, funcs)
	}
	// Token-based fallback analysis
	// collect Owned<T> parameters by name
	owned := map[string]bool{}
	for _, p := range fd.Params {
		if p.Name != "" && strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
			owned[p.Name] = true
		}
	}
	if len(owned) == 0 {
		return diags
	}
	released := map[string]bool{}
	usedAfter := map[string]bool{}
	// helper: process function call at tokens[i] being callee IDENT or receiver IDENT '.' method
	toks := fd.Body
	isReleaser := func(name string) bool {
		switch strings.ToLower(name) {
		case "release", "drop", "free", "dispose":
			return true
		}
		return false
	}
	// parse call args starting at index of '('; returns list of top-level identifier args and end index of ')'
	parseArgs := func(start int) ([]string, int) {
		args := []string{}
		depth := 0
		curIdent := ""
		end := start
		for i := start; i < len(toks); i++ {
			end = i
			tk := toks[i]
			if tk.Kind == tok.LPAREN {
				depth++
				continue
			}
			if tk.Kind == tok.RPAREN {
				depth--
				if depth == 0 {
					break
				}
				continue
			}
			if depth == 1 { // top-level within call
				if tk.Kind == tok.IDENT {
					curIdent = tk.Lexeme
				} else {
					if curIdent != "" {
						args = append(args, curIdent)
						curIdent = ""
					}
				}
				if tk.Kind == tok.COMMA {
					if curIdent != "" {
						args = append(args, curIdent)
						curIdent = ""
					}
				}
			}
		}
		if curIdent != "" {
			args = append(args, curIdent)
		}
		return args, end
	}
	// map for known function signatures
	// scan tokens
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		// use-after-release detection (simple): any occurrence of owned ident after release
		if t.Kind == tok.IDENT && owned[t.Lexeme] && released[t.Lexeme] {
			if !usedAfter[t.Lexeme] {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use of owned value after release/transfer"})
				usedAfter[t.Lexeme] = true
			}
		}
		// function call: IDENT '(' ... ')'
		if t.Kind == tok.IDENT && i+1 < len(toks) && toks[i+1].Kind == tok.LPAREN {
			callee := t.Lexeme
			args, end := parseArgs(i + 1)
			// if callee is known function, transfer ownership for args matching Owned params
			if fd2, ok := funcs[callee]; ok {
				// detect if callee accepts any Owned parameter (position-agnostic fallback)
				calleeHasOwned := false
				for _, p := range fd2.Params {
					if strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
						calleeHasOwned = true
						break
					}
				}
				matched := false
				for idx, a := range args {
					if !owned[a] {
						continue
					}
					// precise positional check
					transferred := false
					if idx < len(fd2.Params) {
						pt := fd2.Params[idx].Type
						if strings.ToLower(pt.Name) == "owned" && len(pt.Args) == 1 {
							transferred = true
						}
					}
					if !transferred && calleeHasOwned {
						transferred = true
					}
					if transferred {
						if released[a] {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release/transfer of owned value"})
						}
						released[a] = true
						matched = true
					}
				}
				if !matched && calleeHasOwned {
					// heuristic fallback: release any single owned param if only one exists
					count := 0
					last := ""
					for name := range owned {
						count++
						last = name
					}
					if count == 1 {
						if released[last] {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release/transfer of owned value"})
						}
						released[last] = true
					}
				}
			}
			// releaser by name
			if isReleaser(callee) {
				releasedAny := false
				for _, a := range args {
					if owned[a] {
						if released[a] {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
						}
						released[a] = true
						releasedAny = true
					}
				}
				if !releasedAny {
					// conservative: release all owned params
					for name := range owned {
						if !released[name] {
							released[name] = true
						}
					}
				}
			}
			i = end // jump to end of call
			continue
		}
		// method call: IDENT '.' IDENT '(' ... ')'
		if t.Kind == tok.IDENT && i+3 < len(toks) && toks[i+1].Kind == tok.DOT && toks[i+2].Kind == tok.IDENT && toks[i+3].Kind == tok.LPAREN {
			recv := t.Lexeme
			mth := toks[i+2].Lexeme
			_, end := parseArgs(i + 3)
			if owned[recv] {
				switch strings.ToLower(mth) {
				case "close", "release", "free", "dispose":
					if released[recv] {
						diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
					}
					released[recv] = true
				}
			}
			i = end
			continue
		}
	}
	// end-of-function: any owned param not released/transferred -> diagnostic
	if !isSinkFunction(fd) {
		for name := range owned {
			if !released[name] {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_OWNED_NOT_RELEASED", Message: "owned value not released or transferred before function end"})
			}
		}
	}
	return diags
}

func analyzeRAIIFromAST(fd astpkg.FuncDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	// owned params
	owned := map[string]bool{}
	for _, p := range fd.Params {
		if p.Name != "" && strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
			owned[p.Name] = true
		}
	}
	if len(owned) == 0 {
		return diags
	}
	released := map[string]bool{}
	deferred := map[string]bool{}
	usedAfter := map[string]bool{}
	// helpers
	isReleaser := func(name string) bool {
		switch strings.ToLower(name) {
		case "release", "drop", "free", "dispose":
			return true
		}
		return false
	}
	var walkExpr func(astpkg.Expr)
	walkExpr = func(e astpkg.Expr) {
		switch v := e.(type) {
		case astpkg.Ident:
			if owned[v.Name] && released[v.Name] && !usedAfter[v.Name] {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use of owned value after release/transfer"})
				usedAfter[v.Name] = true
			}
		case astpkg.UnaryExpr:
			walkExpr(v.X)
		case astpkg.SelectorExpr:
			// receiver use counts as use-after
			if id, ok := v.X.(astpkg.Ident); ok {
				if owned[id.Name] {
					if strings.EqualFold(v.Sel, "close") || strings.EqualFold(v.Sel, "release") || strings.EqualFold(v.Sel, "free") || strings.EqualFold(v.Sel, "dispose") {
						if released[id.Name] || deferred[id.Name] {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
						}
						released[id.Name] = true
					} else if released[id.Name] && !usedAfter[id.Name] {
						diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use of owned value after release/transfer"})
						usedAfter[id.Name] = true
					}
				}
			}
		case astpkg.CallExpr:
			// function/method calls
			switch f := v.Fun.(type) {
			case astpkg.Ident:
				name := f.Name
				// releaser by name
				if isReleaser(name) {
					mark := false
					for _, a := range v.Args {
						if id, ok := a.(astpkg.Ident); ok && owned[id.Name] {
							if released[id.Name] || deferred[id.Name] {
								diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
							}
							released[id.Name] = true
							mark = true
						}
					}
					if !mark {
						// conservative: release single owned param if only one
						if len(owned) == 1 {
							for k := range owned {
								released[k] = true
							}
						}
					}
				}
				// Always visit args to catch nested calls (e.g., mutate(release(x)))
				for _, a := range v.Args {
					walkExpr(a)
				}
				// transfer via owned parameter position
				if callee, ok := funcs[name]; ok {
					calleeHasOwned := false
					for _, p := range callee.Params {
						if strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
							calleeHasOwned = true
							break
						}
					}
					matched := false
					for idx, a := range v.Args {
						if id, ok := a.(astpkg.Ident); ok && owned[id.Name] {
							if idx < len(callee.Params) {
								pt := callee.Params[idx].Type
								if strings.ToLower(pt.Name) == "owned" && len(pt.Args) == 1 {
									released[id.Name] = true
									matched = true
								}
							}
						}
					}
					if !matched && calleeHasOwned && len(owned) == 1 {
						for k := range owned {
							released[k] = true
						}
					}
				}
			case astpkg.SelectorExpr:
				// Method call: receiver.method(args)
				// Treat known releaser methods as release operations.
				if id, ok := f.X.(astpkg.Ident); ok && owned[id.Name] {
					if strings.EqualFold(f.Sel, "close") || strings.EqualFold(f.Sel, "release") || strings.EqualFold(f.Sel, "free") || strings.EqualFold(f.Sel, "dispose") {
						if released[id.Name] || deferred[id.Name] {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
						}
						released[id.Name] = true
					} else if released[id.Name] && !usedAfter[id.Name] {
						diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_USE_AFTER_RELEASE", Message: "use of owned value after release/transfer"})
						usedAfter[id.Name] = true
					}
				}
				// Visit args for use-after detection in their subtrees
				for _, a := range v.Args {
					walkExpr(a)
				}
			default:
				for _, a := range v.Args {
					walkExpr(a)
				}
			}
		}
	}
	var walkStmt func(astpkg.Stmt)
	walkStmt = func(s astpkg.Stmt) {
		switch v := s.(type) {
		case astpkg.AssignStmt:
			walkExpr(v.LHS)
			walkExpr(v.RHS)
		case astpkg.ExprStmt:
			walkExpr(v.X)
		case astpkg.DeferStmt:
			// analyze deferred call specially: schedule release/transfer at end
			if ce, ok := v.X.(astpkg.CallExpr); ok {
				switch f := ce.Fun.(type) {
				case astpkg.Ident:
					name := f.Name
					// known releaser by name
					if isReleaser(name) {
						mark := false
						for _, a := range ce.Args {
							if id, ok := a.(astpkg.Ident); ok && owned[id.Name] {
								if released[id.Name] || deferred[id.Name] {
									diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
								}
								deferred[id.Name] = true
								mark = true
							}
						}
						if !mark {
							if len(owned) == 1 {
								for k := range owned {
									if released[k] {
										diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
									}
									deferred[k] = true
								}
							}
						}
					}
					// Transfer semantics for known functions with Owned params: treat as release at end
					if callee, ok := funcs[name]; ok {
						calleeHasOwned := false
						for _, p := range callee.Params {
							if strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
								calleeHasOwned = true
								break
							}
						}
						matched := false
						for idx, a := range ce.Args {
							if id, ok := a.(astpkg.Ident); ok && owned[id.Name] {
								if idx < len(callee.Params) {
									pt := callee.Params[idx].Type
									if strings.ToLower(pt.Name) == "owned" && len(pt.Args) == 1 {
										if released[id.Name] || deferred[id.Name] {
											diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release/transfer of owned value"})
										}
										deferred[id.Name] = true
										matched = true
									}
								}
							}
						}
						if !matched && calleeHasOwned && len(owned) == 1 {
							for k := range owned {
								if released[k] || deferred[k] {
									diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release/transfer of owned value"})
								}
								deferred[k] = true
							}
						}
					}
				case astpkg.SelectorExpr:
					// receiver.method(...)
					if id, ok := f.X.(astpkg.Ident); ok && owned[id.Name] {
						if strings.EqualFold(f.Sel, "close") || strings.EqualFold(f.Sel, "release") || strings.EqualFold(f.Sel, "free") || strings.EqualFold(f.Sel, "dispose") {
							if released[id.Name] || deferred[id.Name] {
								diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_DOUBLE_RELEASE", Message: "double release of owned value"})
							}
							deferred[id.Name] = true
						}
					}
				default:
					// nothing
				}
			}
		case astpkg.MutBlockStmt:
			for _, ss := range v.Body.Stmts {
				walkStmt(ss)
			}
		case astpkg.BlockStmt:
			for _, ss := range v.Stmts {
				walkStmt(ss)
			}
		}
	}
	for _, s := range fd.BodyStmts {
		walkStmt(s)
	}
	if !isSinkFunction(fd) {
		for name := range owned {
			if !(released[name] || deferred[name]) {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RAII_OWNED_NOT_RELEASED", Message: "owned value not released or transferred before function end"})
			}
		}
	}
	return diags
}

// analyzeMemoryDomains detects cross-domain references between Event payload and State
// specifically: *st = &ev (assigning address of event into state) emits E_CROSS_DOMAIN_REF.
// (Removed old duplicate analyzeMemoryDomains; consolidated above.)

// isSinkFunction heuristically identifies helper functions that consume Owned<T>
// parameters and are not subject to RAII not-released enforcement. Criteria:
// - Function is not a worker signature; and
// - Has at least one parameter of type Owned<…>; and
// - Has exactly one parameter (common sink pattern).
func isSinkFunction(fd astpkg.FuncDecl) bool {
	if isWorkerSignature(fd) {
		return false
	}
	hasOwned := false
	for _, p := range fd.Params {
		if strings.EqualFold(p.Type.Name, "owned") && len(p.Type.Args) == 1 {
			hasOwned = true
			break
		}
	}
	if !hasOwned {
		return false
	}
	if len(fd.Params) == 1 {
		return true
	}
	return false
}

// analyzeEventTypeFlow ensures that the payload type of upstream worker outputs
// matches the Event<T> input payload type of downstream workers for each step
// transition in a pipeline's normal path.
func analyzeEventTypeFlow(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	// helper to get worker output payload type string
	workerOut := func(name string) (string, bool) {
		fd, ok := funcs[name]
		if !ok || len(fd.Result) != 1 {
			return "", false
		}
		r := fd.Result[0]
		if r.Name == "Event" && len(r.Args) == 1 {
			return typeRefToString(r.Args[0]), true
		}
		if r.Name == "Error" && len(r.Args) == 1 {
			return "", false
		} // skip error output in normal flow
		return "", false
	}
	// helper to get worker input payload type string
	workerIn := func(name string) (string, bool) {
		fd, ok := funcs[name]
		if !ok || len(fd.Params) < 2 {
			return "", false
		}
		p2 := fd.Params[1].Type
		if p2.Name == "Event" && len(p2.Args) == 1 {
			return typeRefToString(p2.Args[0]), true
		}
		return "", false
	}
	for i := 1; i < len(pd.Steps); i++ {
		prev := pd.Steps[i-1]
		next := pd.Steps[i]
		var outs []string
		for _, w := range prev.Workers {
			if t, ok := workerOut(w.Name); ok {
				outs = append(outs, t)
			}
		}
		if len(outs) == 0 {
			continue
		}
		var ins []string
		for _, w := range next.Workers {
			if t, ok := workerIn(w.Name); ok {
				ins = append(ins, t)
			}
		}
		// If no downstream worker inputs (e.g., collect/egress), skip
		if len(ins) == 0 {
			continue
		}
		// All combinations must match (with conservative generic compatibility)
		for _, o := range outs {
			for _, in := range ins {
				if o == in {
					continue
				}
				if isGenericEvent(o) && isGenericEvent(in) {
					continue
				}
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EVENT_TYPE_FLOW", Message: "event payload type mismatch between upstream worker output and downstream input"})
				// one diag per boundary is enough
				goto nextStep
			}
		}
	nextStep:
		continue
	}
	return diags
}

func isGenericEvent(s string) bool {
	if len(s) < 9 {
		return false
	}
	if !strings.HasPrefix(s, "Event<") || !strings.HasSuffix(s, ">") {
		return false
	}
	inner := s[len("Event<") : len(s)-1]
	if len(inner) == 1 {
		b := inner[0]
		return b >= 'A' && b <= 'Z'
	}
	return false
}

// analyzeImperativeTypes performs minimal type checking over function bodies by
// scanning tokens and leveraging parameter types as an environment.
// Supported checks:
//   - E_DEREF_TYPE: '*' applied to non-pointer parameter identifier.
//   - E_ASSIGN_TYPE_MISMATCH: for simple forms `x = y`, `*p = y`, `x = &y` when
//     both sides resolve to known types from parameters or literals.
func analyzeImperativeTypes(fd astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if len(fd.Body) == 0 {
		return diags
	}
	if len(fd.BodyStmts) > 0 {
		return analyzeImperativeTypesFromAST(fd)
	}
	// env maps parameter identifiers to their TypeRef
	env := map[string]astpkg.TypeRef{}
	for _, p := range fd.Params {
		if p.Name != "" {
			env[p.Name] = p.Type
		}
	}
	// helpers
	typeStr := func(tr astpkg.TypeRef) string { return typeRefToString(tr) }
	// resolve expression type starting at token index i; returns type string and ok
	// Supports IDENT, '*' IDENT, '&' IDENT, STRING→string, NUMBER→int
	resolve := func(toks []tok.Token, i int) (string, bool, bool) {
		// third return is "hardError" indicator already emitted; used to avoid double-diags
		if i >= len(toks) {
			return "", false, false
		}
		switch toks[i].Kind {
		case tok.IDENT:
			if tr, ok := env[toks[i].Lexeme]; ok {
				return typeStr(tr), true, false
			}
			return "", false, false
		case tok.STAR, tok.AMP:
			// no raw pointer semantics in AMI
			return "", false, false
		case tok.STRING:
			return "string", true, false
		case tok.NUMBER:
			return "int", true, false
		default:
			return "", false, false
		}
	}
	toks := fd.Body
	for i := 0; i < len(toks); i++ {
		// Simple assignment patterns: [IDENT|* IDENT] '=' <expr>
		if toks[i].Kind == tok.ASSIGN {
			// LHS type
			var lhs string
			var okL bool
			if !okL && i-1 >= 0 && toks[i-1].Kind == tok.IDENT {
				if tr, ok := env[toks[i-1].Lexeme]; ok {
					lhs, okL = typeStr(tr), true
				}
			}
			// RHS type
			rhs, okR, _ := resolve(toks, i+1)
			if okL && okR && lhs != rhs {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ASSIGN_TYPE_MISMATCH", Message: "assignment type mismatch: " + lhs + " != " + rhs})
			}
		}
	}
	return diags
}

func analyzeImperativeTypesFromAST(fd astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	env := map[string]astpkg.TypeRef{}
	for _, p := range fd.Params {
		if p.Name != "" {
			env[p.Name] = p.Type
		}
	}
	typeStr := func(tr astpkg.TypeRef) string { return typeRefToString(tr) }
	// single-letter type variable helper & unifier
	isTypeVar := func(name string) bool { return len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z' }
	var unify func(want, got astpkg.TypeRef, subst map[string]astpkg.TypeRef) bool
	unify = func(want, got astpkg.TypeRef, subst map[string]astpkg.TypeRef) bool {
		if want.Ptr != got.Ptr || want.Slice != got.Slice {
			return false
		}
		if isTypeVar(want.Name) && len(want.Args) == 0 {
			if b, ok := subst[want.Name]; ok {
				return typeStr(b) == typeStr(got)
			}
			subst[want.Name] = got
			return true
		}
		if strings.ToLower(want.Name) != strings.ToLower(got.Name) {
			return false
		}
		if len(want.Args) != len(got.Args) {
			return false
		}
		for i := range want.Args {
			if !unify(want.Args[i], got.Args[i], subst) {
				return false
			}
		}
		return true
	}
	var exprType func(astpkg.Expr, bool) (astpkg.TypeRef, bool)
	exprType = func(e astpkg.Expr, isLHS bool) (astpkg.TypeRef, bool) {
		switch v := e.(type) {
		case astpkg.Ident:
			if tr, ok := env[v.Name]; ok {
				return tr, true
			}
			return astpkg.TypeRef{}, false
		case astpkg.ContainerLit:
			switch v.Kind {
			case "slice", "set":
				// element type is TypeArgs[0] if provided; otherwise infer from elements
				var elemT astpkg.TypeRef
				hasType := len(v.TypeArgs) == 1
				if hasType {
					elemT = v.TypeArgs[0]
				}
				subst := map[string]astpkg.TypeRef{}
				for _, el := range v.Elems {
					et, ok := exprType(el, false)
					if !ok {
						continue
					}
					if hasType {
						if !unify(elemT, et, subst) {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "container element type mismatch"})
						}
					} else {
						// first element determines type; subsequent must match
						if elemT.Name == "" {
							elemT = et
						} else if typeStr(elemT) != typeStr(et) {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "container element types must match"})
						}
					}
				}
				if v.Kind == "slice" {
					return astpkg.TypeRef{Name: "slice", Args: []astpkg.TypeRef{elemT}}, true
				}
				return astpkg.TypeRef{Name: "set", Args: []astpkg.TypeRef{elemT}}, true
			case "map":
				var keyT, valT astpkg.TypeRef
				hasTypes := len(v.TypeArgs) == 2
				if hasTypes {
					keyT, valT = v.TypeArgs[0], v.TypeArgs[1]
				}
				substK := map[string]astpkg.TypeRef{}
				substV := map[string]astpkg.TypeRef{}
				for _, kv := range v.MapElems {
					kt, okk := exprType(kv.Key, false)
					if !okk {
						continue
					}
					vt, okv := exprType(kv.Value, false)
					if !okv {
						continue
					}
					if hasTypes {
						if !unify(keyT, kt, substK) || !unify(valT, vt, substV) {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map element type mismatch"})
						}
					} else {
						if keyT.Name == "" {
							keyT = kt
						} else if typeStr(keyT) != typeStr(kt) {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map key types must match"})
						}
						if valT.Name == "" {
							valT = vt
						} else if typeStr(valT) != typeStr(vt) {
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "map value types must match"})
						}
					}
				}
				return astpkg.TypeRef{Name: "map", Args: []astpkg.TypeRef{keyT, valT}}, true
			}
			return astpkg.TypeRef{}, false
		case astpkg.UnaryExpr:
			// '*' is a marker: yields the same type as its operand on LHS
			if v.Op == "*" {
				t, ok := exprType(v.X, isLHS)
				if !ok {
					return astpkg.TypeRef{}, false
				}
				if isLHS && t.Ptr {
					t.Ptr = false
				}
				return t, true
			}
			// '&' is prohibited; treat as unknown here
			return astpkg.TypeRef{}, false
		case astpkg.BinaryExpr:
			// Determine operand types
			lt, lok := exprType(v.X, false)
			rt, rok := exprType(v.Y, false)
			if !(lok && rok) {
				return astpkg.TypeRef{}, false
			}
			// arithmetic operators
			switch v.Op {
			case "+", "-", "*", "/", "%":
				if strings.ToLower(lt.Name) == "int" && strings.ToLower(rt.Name) == "int" {
					return astpkg.TypeRef{Name: "int"}, true
				}
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "arithmetic operands must be numeric (int)", Pos: &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
				return astpkg.TypeRef{}, false
			case "==", "!=":
				if typeStr(lt) == typeStr(rt) {
					return astpkg.TypeRef{Name: "bool"}, true
				}
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "comparison operands must have identical types", Pos: &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
				return astpkg.TypeRef{}, false
			case "<", "<=", ">", ">=":
				if strings.ToLower(lt.Name) == "int" && strings.ToLower(rt.Name) == "int" {
					return astpkg.TypeRef{Name: "bool"}, true
				}
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "ordering comparisons require numeric (int) operands", Pos: &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}})
				return astpkg.TypeRef{}, false
			}
			return astpkg.TypeRef{}, false
		case astpkg.BasicLit:
			switch v.Kind {
			case "string":
				return astpkg.TypeRef{Name: "string"}, true
			case "number":
				return astpkg.TypeRef{Name: "int"}, true
			}
			return astpkg.TypeRef{}, false
		default:
			return astpkg.TypeRef{}, false
		}
	}
	var walkStmt func(astpkg.Stmt)
	walkStmt = func(s astpkg.Stmt) {
		if as, ok := s.(astpkg.AssignStmt); ok {
			lt, lok := exprType(as.LHS, true)
			rt, rok := exprType(as.RHS, false)
			if lok && rok {
				subst := map[string]astpkg.TypeRef{}
				if !unify(lt, rt, subst) {
					d := diag.Diagnostic{Level: diag.Error, Code: "E_ASSIGN_TYPE_MISMATCH", Message: "assignment type mismatch: " + typeStr(lt) + " != " + typeStr(rt)}
					d.Pos = &srcset.Position{Line: as.Pos.Line, Column: as.Pos.Column, Offset: as.Pos.Offset}
					diags = append(diags, d)
				}
			}
			return
		}
		switch v := s.(type) {
		case astpkg.VarDeclStmt:
			// Handle var decls: var name [Type] [= init]
			if v.Type.Name != "" && v.Init != nil {
				rt, rok := exprType(v.Init, false)
				if rok {
					subst := map[string]astpkg.TypeRef{}
					if !unify(v.Type, rt, subst) {
						d := diag.Diagnostic{Level: diag.Error, Code: "E_ASSIGN_TYPE_MISMATCH", Message: "var init type mismatch: " + typeStr(v.Type) + " != " + typeStr(rt)}
						d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
						diags = append(diags, d)
					}
				}
				env[v.Name] = v.Type
			} else if v.Type.Name != "" && v.Init == nil {
				env[v.Name] = v.Type
			} else if v.Type.Name == "" && v.Init != nil {
				if rt, rok := exprType(v.Init, false); rok {
					env[v.Name] = rt
				} else {
					d := diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_UNINFERRED", Message: "cannot infer variable type from initializer"}
					d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
					diags = append(diags, d)
				}
			} else {
				// no type and no init
				d := diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_UNINFERRED", Message: "variable declaration missing type and initializer"}
				d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
				diags = append(diags, d)
			}
		case astpkg.ExprStmt:
			// type derivation for expression statements (ensures BinaryExpr checks run)
			_, _ = exprType(v.X, false)
		case astpkg.ReturnStmt:
			// validate return types
			if len(fd.Result) == 0 {
				if len(v.Results) > 0 {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "function has no return values"})
				}
				return
			}
			if len(fd.Result) != len(v.Results) {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return value count does not match function result arity"})
				return
			}
			// unify each returned expression against declared result
			for i, rexpr := range v.Results {
				rt, rok := exprType(rexpr, false)
				if !rok {
					continue
				}
				want := fd.Result[i]
				subst := map[string]astpkg.TypeRef{}
				if !unify(want, rt, subst) {
					d := diag.Diagnostic{Level: diag.Error, Code: "E_RETURN_TYPE_MISMATCH", Message: "return type mismatch: got " + typeStr(rt) + ", want " + typeStr(want)}
					d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
					diags = append(diags, d)
					continue
				}
				// if any substitution binds to a type variable (not concrete), report uninferred
				for _, b := range subst {
					if len(b.Name) == 1 && b.Name[0] >= 'A' && b.Name[0] <= 'Z' {
						d := diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_UNINFERRED", Message: "return type contains uninferred type variables"}
						d.Pos = &srcset.Position{Line: v.Pos.Line, Column: v.Pos.Column, Offset: v.Pos.Offset}
						diags = append(diags, d)
						break
					}
				}
			}
		case astpkg.BlockStmt:
			for _, ss := range v.Stmts {
				walkStmt(ss)
			}
		}
	}
	for _, s := range fd.BodyStmts {
		walkStmt(s)
	}
	return diags
}

// analyzeOperators scans tokens to validate basic arithmetic and comparison operand types.
// Arithmetic: +,-,*,/,% expect numeric (int) on both sides.
// Comparison: ==,!= require same types; <,<=,>,>= allowed for int (and strings not yet guaranteed).
func analyzeOperators(fd astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	// If AST is available, operator checks are handled via exprType on BinaryExpr
	if len(fd.BodyStmts) > 0 {
		return diags
	}
	if len(fd.Body) == 0 {
		return diags
	}
	// build env from params
	env := map[string]astpkg.TypeRef{}
	for _, p := range fd.Params {
		if p.Name != "" {
			env[p.Name] = p.Type
		}
	}
	// helpers
	resolve := func(t tok.Token) (astpkg.TypeRef, bool) {
		switch t.Kind {
		case tok.IDENT:
			if tr, ok := env[t.Lexeme]; ok {
				return tr, true
			}
			return astpkg.TypeRef{}, false
		case tok.NUMBER:
			return astpkg.TypeRef{Name: "int"}, true
		case tok.STRING:
			return astpkg.TypeRef{Name: "string"}, true
		default:
			return astpkg.TypeRef{}, false
		}
	}
	prevMeaningful := func(toks []tok.Token, i int) (tok.Token, bool) {
		for j := i - 1; j >= 0; j-- {
			if toks[j].Kind != tok.SEMI {
				return toks[j], true
			}
		}
		return tok.Token{}, false
	}
	nextMeaningful := func(toks []tok.Token, i int) (tok.Token, bool) {
		for j := i + 1; j < len(toks); j++ {
			if toks[j].Kind != tok.SEMI {
				return toks[j], true
			}
		}
		return tok.Token{}, false
	}
	isArithmetic := func(k tok.Kind) bool {
		return k == tok.PLUS || k == tok.MINUS || k == tok.SLASH || k == tok.PERCENT || k == tok.STAR
	}
	isComparison := func(k tok.Kind) bool {
		return k == tok.EQ || k == tok.NEQ || k == tok.LT || k == tok.LTE || k == tok.GT || k == tok.GTE
	}

	toks := fd.Body
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if isArithmetic(t.Kind) || isComparison(t.Kind) {
			// treat '*' as multiplication only when not the LHS mutation marker: pattern '*' IDENT '='
			if t.Kind == tok.STAR {
				if i+2 < len(toks) && toks[i+1].Kind == tok.IDENT && toks[i+2].Kind == tok.ASSIGN {
					continue
				}
			}
			lTok, okL := prevMeaningful(toks, i)
			rTok, okR := nextMeaningful(toks, i)
			if !okL || !okR {
				continue
			}
			// ensure these look like operands
			lt, lOk := resolve(lTok)
			rt, rOk := resolve(rTok)
			if !(lOk && rOk) {
				continue
			}
			// arithmetic checks
			if isArithmetic(t.Kind) {
				if !(strings.ToLower(lt.Name) == "int" && strings.ToLower(rt.Name) == "int") {
					diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "arithmetic operands must be numeric (int)"})
				}
				continue
			}
			// comparison checks
			if isComparison(t.Kind) {
				if t.Kind == tok.EQ || t.Kind == tok.NEQ {
					if typeRefToString(lt) != typeRefToString(rt) {
						diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "comparison operands must have identical types"})
					}
				} else {
					// ordering comparisons limited to int for now
					if strings.ToLower(lt.Name) != "int" || strings.ToLower(rt.Name) != "int" {
						diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_MISMATCH", Message: "ordering comparisons require numeric (int) operands"})
					}
				}
			}
		}
	}
	return diags
}

// analyzeCallTypes walks function bodies (AST form) and validates that
// argument expressions passed to known local functions are type-compatible
// with their parameter types. It accepts simple generic forms where the
// callee's parameter uses a single-letter type variable (e.g., Event<T>),
// unifying it against the actual argument type.
func analyzeCallTypes(fd astpkg.FuncDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	if len(fd.BodyStmts) == 0 {
		return diags
	}
	// environment: parameter identifier -> TypeRef
	env := map[string]astpkg.TypeRef{}
	for _, p := range fd.Params {
		if p.Name != "" {
			env[p.Name] = p.Type
		}
	}

	// infer expression type where possible
	var exprType func(astpkg.Expr) (astpkg.TypeRef, bool)
	exprType = func(e astpkg.Expr) (astpkg.TypeRef, bool) {
		switch v := e.(type) {
		case astpkg.Ident:
			if tr, ok := env[v.Name]; ok {
				return tr, true
			}
			return astpkg.TypeRef{}, false
		case astpkg.ContainerLit:
			switch v.Kind {
			case "slice":
				var elem astpkg.TypeRef
				if len(v.TypeArgs) == 1 {
					elem = v.TypeArgs[0]
				} else if len(v.Elems) > 0 {
					if t, ok := exprType(v.Elems[0]); ok {
						elem = t
					}
				}
				return astpkg.TypeRef{Name: "slice", Args: []astpkg.TypeRef{elem}}, true
			case "set":
				var elem astpkg.TypeRef
				if len(v.TypeArgs) == 1 {
					elem = v.TypeArgs[0]
				} else if len(v.Elems) > 0 {
					if t, ok := exprType(v.Elems[0]); ok {
						elem = t
					}
				}
				return astpkg.TypeRef{Name: "set", Args: []astpkg.TypeRef{elem}}, true
			case "map":
				var kt, vt astpkg.TypeRef
				if len(v.TypeArgs) == 2 {
					kt, vt = v.TypeArgs[0], v.TypeArgs[1]
				} else if len(v.MapElems) > 0 {
					if t1, ok := exprType(v.MapElems[0].Key); ok {
						kt = t1
					}
					if t2, ok := exprType(v.MapElems[0].Value); ok {
						vt = t2
					}
				}
				return astpkg.TypeRef{Name: "map", Args: []astpkg.TypeRef{kt, vt}}, true
			}
			return astpkg.TypeRef{}, false
		case astpkg.BasicLit:
			if v.Kind == "string" {
				return astpkg.TypeRef{Name: "string"}, true
			}
			if v.Kind == "number" {
				return astpkg.TypeRef{Name: "int"}, true
			}
			return astpkg.TypeRef{}, false
		case astpkg.UnaryExpr:
			if v.Op == "*" {
				return exprType(v.X)
			}
			return astpkg.TypeRef{}, false
		case astpkg.BinaryExpr:
			// derive simple operator result types
			lt, lok := exprType(v.X)
			rt, rok := exprType(v.Y)
			if !(lok && rok) {
				return astpkg.TypeRef{}, false
			}
			switch v.Op {
			case "+", "-", "*", "/", "%":
				if strings.ToLower(lt.Name) == "int" && strings.ToLower(rt.Name) == "int" {
					return astpkg.TypeRef{Name: "int"}, true
				}
				return astpkg.TypeRef{}, false
			case "==", "!=", "<", "<=", ">", ">=":
				return astpkg.TypeRef{Name: "bool"}, true
			}
			return astpkg.TypeRef{}, false
		case astpkg.CallExpr:
			// If callee is an identifier and we know its declaration, and it returns exactly one type, propagate that.
			switch c := v.Fun.(type) {
			case astpkg.Ident:
				if decl, ok := funcs[c.Name]; ok {
					if len(decl.Result) == 1 {
						return decl.Result[0], true
					}
				}
			case astpkg.SelectorExpr:
				// Unknown for now
			}
			return astpkg.TypeRef{}, false
		default:
			return astpkg.TypeRef{}, false
		}
	}

	// Simple structural unify with single-letter type variables
	type substMap = map[string]astpkg.TypeRef
	var isTypeVar = func(name string) bool {
		if len(name) != 1 {
			return false
		}
		b := name[0]
		return b >= 'A' && b <= 'Z'
	}
	var unify func(want, got astpkg.TypeRef, subst substMap) bool
	unify = func(want, got astpkg.TypeRef, subst substMap) bool {
		// pointer and slice flags must match exactly
		if want.Ptr != got.Ptr || want.Slice != got.Slice {
			return false
		}
		// Generic variable binding
		if isTypeVar(want.Name) && len(want.Args) == 0 {
			if bound, ok := subst[want.Name]; ok {
				return typeRefToString(bound) == typeRefToString(got)
			}
			subst[want.Name] = got
			return true
		}
		if strings.ToLower(want.Name) != strings.ToLower(got.Name) {
			return false
		}
		if len(want.Args) != len(got.Args) {
			return false
		}
		for i := range want.Args {
			if !unify(want.Args[i], got.Args[i], subst) {
				return false
			}
		}
		return true
	}

	// Walk statements (track local var decls) and check calls
	var walkStmt func(astpkg.Stmt)
	var walkExpr func(astpkg.Expr)
	walkExpr = func(e astpkg.Expr) {
		switch v := e.(type) {
		case astpkg.CallExpr:
			// Only check calls to known local functions
			if id, ok := v.Fun.(astpkg.Ident); ok {
				if decl, ok2 := funcs[id.Name]; ok2 {
					// Arity check
					if len(v.Args) != len(decl.Params) {
						diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CALL_ARITY_MISMATCH", Message: "function call arity mismatch"})
						// still try to compare what we can
					}
					n := len(v.Args)
					if len(decl.Params) < n {
						n = len(decl.Params)
					}
					// reset substitution per call
					subst := make(substMap)
					for i := 0; i < n; i++ {
						at, aok := exprType(v.Args[i])
						if !aok {
							// if parameter expects a type variable (directly or as generic arg), mark ambiguous
							pt := decl.Params[i].Type
							if hasTypeVar(pt) {
								diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_TYPE_AMBIGUOUS", Message: "cannot infer generic type argument from call site"})
							}
							continue
						}
						pt := decl.Params[i].Type
						if !unify(pt, at, subst) {
							msg := "call argument type mismatch: got " + typeRefToString(at) + ", want " + typeRefToString(pt)
							diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CALL_ARG_TYPE_MISMATCH", Message: msg})
						}
					}
				}
			} else {
				// nested selector/unknown callee: still walk args
			}
			for _, a := range v.Args {
				walkExpr(a)
			}
		case astpkg.UnaryExpr:
			walkExpr(v.X)
		case astpkg.SelectorExpr:
			walkExpr(v.X)
		case astpkg.Expr:
			// nothing else to walk
			_ = v
		}
	}
	walkStmt = func(s astpkg.Stmt) {
		switch v := s.(type) {
		case astpkg.VarDeclStmt:
			if v.Type.Name != "" {
				env[v.Name] = v.Type
			} else if v.Init != nil {
				if t, ok := exprType(v.Init); ok {
					env[v.Name] = t
				}
			}
		case astpkg.AssignStmt:
			walkExpr(v.LHS)
			walkExpr(v.RHS)
		case astpkg.ExprStmt:
			walkExpr(v.X)
		case astpkg.BlockStmt:
			for _, ss := range v.Stmts {
				walkStmt(ss)
			}
		case astpkg.DeferStmt:
			if v.X != nil {
				walkExpr(v.X)
			}
		case astpkg.MutBlockStmt:
			for _, ss := range v.Body.Stmts {
				walkStmt(ss)
			}
		case astpkg.ReturnStmt:
			for _, r := range v.Results {
				walkExpr(r)
			}
		default:
			// nothing
		}
	}
	for _, s := range fd.BodyStmts {
		walkStmt(s)
	}
	return diags
}

func hasTypeVar(t astpkg.TypeRef) bool {
	if len(t.Name) == 1 && t.Name[0] >= 'A' && t.Name[0] <= 'Z' {
		return true
	}
	for _, a := range t.Args {
		if hasTypeVar(a) {
			return true
		}
	}
	return false
}

func isWorkerSignature(fd astpkg.FuncDecl) bool {
	// params: (Context, Event<T>, State)
	if len(fd.Params) != 3 {
		return false
	}
	p1 := fd.Params[0].Type
	p2 := fd.Params[1].Type
	p3 := fd.Params[2].Type
	if !(p1.Name == "Context" && !p1.Ptr && !p1.Slice) {
		return false
	}
	if !(p2.Name == "Event" && len(p2.Args) == 1 && !p2.Ptr) {
		return false
	}
	if !(p3.Name == "State") {
		return false
	}
	// results: exactly one of Event<U>, []Event<U>, Error<E>
	if len(fd.Result) != 1 {
		return false
	}
	r := fd.Result[0]
	switch {
	case r.Name == "Event" && len(r.Args) == 1 && !r.Slice:
		return true
	case r.Name == "Event" && len(r.Args) == 1 && r.Slice:
		return true
	case r.Name == "Error" && len(r.Args) == 1:
		return true
	default:
		return false
	}
}

// analyzeMutability: implemented above. This section intentionally left
// without duplication to avoid redeclaration.

// analyzeIOPermissions enforces that only ingress/egress nodes may perform I/O
// when step arguments indicate I/O usage via simple attributes.
// Detection rules (scaffold):
// - Any argument starting with "io=" (e.g., io=read, io=write)
// - Any argument containing "io.read(" or "io.write("
// These forms are only allowed on ingress/egress. Others emit E_IO_PERMISSION.
func analyzeIOPermissions(pd astpkg.PipelineDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	isIOArg := func(s string) bool {
		s = strings.TrimSpace(s)
		if s == "" {
			return false
		}
		if strings.HasPrefix(s, "io=") {
			return true
		}
		if strings.Contains(s, "io.read(") || strings.Contains(s, "io.write(") {
			return true
		}
		return false
	}
	check := func(name string, args []string) {
		n := strings.ToLower(name)
		allowed := (n == "ingress" || n == "egress")
		if allowed {
			return
		}
		for _, a := range args {
			if isIOArg(a) {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_IO_PERMISSION", Message: "I/O operations are only allowed in ingress/egress nodes"})
				break
			}
		}
	}
	for _, st := range pd.Steps {
		check(st.Name, st.Args)
	}
	for _, st := range pd.ErrorSteps {
		check(st.Name, st.Args)
	}
	return diags
}

// analyzeEdgeTypeSafety validates that declared edge `type=` matches the
// upstream worker output payload type for each step.
func analyzeEdgeTypeSafety(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
	var diags []diag.Diagnostic
	// helper to get worker result payload type string
	workerOut := func(name string) (string, bool) {
		fd, ok := funcs[name]
		if !ok {
			return "", false
		}
		if len(fd.Result) != 1 {
			return "", false
		}
		r := fd.Result[0]
		// Event<U> or []Event<U>
		if r.Name == "Event" && len(r.Args) == 1 {
			return typeRefToString(fd.Result[0].Args[0]), true
		}
		if r.Name == "Error" && len(r.Args) == 1 {
			return typeRefToString(fd.Result[0].Args[0]), true
		}
		return "", false
	}
	// Compare step i edge type to previous step workers' outputs
	for i := range pd.Steps {
		st := pd.Steps[i]
		spec, ok := parseEdgeSpecFromArgs(st.Args)
		if !ok {
			continue
		}
		// Extract declared type from spec
		var declared string
		switch v := spec.(type) {
		case fifoSpec:
			declared = v.Type
		case lifoSpec:
			declared = v.Type
		case pipeSpec:
			declared = v.Type
		}
		if declared == "" {
			continue
		}
		// ensure previous step exists
		if i == 0 {
			continue
		}
		prev := pd.Steps[i-1]
		// Gather all worker outputs on previous step
		var outs []string
		for _, w := range prev.Workers {
			if t, ok := workerOut(w.Name); ok {
				outs = append(outs, t)
			}
		}
		// If there were no workers on previous step (e.g., Ingress), skip
		if len(outs) == 0 {
			continue
		}
		// All outputs must match declared type
		for _, t := range outs {
			if t != declared {
				diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_TYPE_MISMATCH", Message: "edge type does not match upstream worker output payload"})
				break
			}
		}
	}
	return diags
}

// analyzeCycles builds a graph of pipeline→pipeline references via edge.Pipeline
// and emits E_CYCLE_DETECTED when a cycle is present unless a `#pragma cycle allow`
// directive is set at file level.
func analyzeCycles(f *astpkg.File) []diag.Diagnostic {
	var diags []diag.Diagnostic
	allow := false
	for _, d := range f.Directives {
		if strings.ToLower(d.Name) == "cycle" && strings.Contains(strings.ToLower(d.Payload), "allow") {
			allow = true
			break
		}
	}
	if allow {
		return diags
	}
	// collect pipelines and edges
	names := []string{}
	idx := map[string]int{}
	for _, n := range f.Decls {
		if p, ok := n.(astpkg.PipelineDecl); ok {
			idx[p.Name] = len(names)
			names = append(names, p.Name)
		}
	}
	g := make([][]int, len(names))
	addEdge := func(from string, to string) {
		i, ok1 := idx[from]
		j, ok2 := idx[to]
		if ok1 && ok2 {
			g[i] = append(g[i], j)
		}
	}
	for _, n := range f.Decls {
		p, ok := n.(astpkg.PipelineDecl)
		if !ok {
			continue
		}
		scan := func(args []string) {
			if spec, ok := parseEdgeSpecFromArgs(args); ok {
				if v, ok2 := spec.(pipeSpec); ok2 {
					if v.Name != "" {
						addEdge(p.Name, v.Name)
					}
				}
			}
		}
		for _, st := range p.Steps {
			scan(st.Args)
		}
		for _, st := range p.ErrorSteps {
			scan(st.Args)
		}
	}
	// detect cycles via DFS
	visited := make([]int, len(names)) // 0=unvisited,1=visiting,2=done
	var dfs func(int) bool
	dfs = func(u int) bool {
		visited[u] = 1
		for _, v := range g[u] {
			if visited[v] == 1 {
				return true
			}
			if visited[v] == 0 && dfs(v) {
				return true
			}
		}
		visited[u] = 2
		return false
	}
	for i := range names {
		if visited[i] == 0 && dfs(i) {
			diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CYCLE_DETECTED", Message: "pipeline graph contains a cycle; add `#pragma cycle allow` with anti-deadlock strategy to permit"})
			break
		}
	}
	return diags
}
