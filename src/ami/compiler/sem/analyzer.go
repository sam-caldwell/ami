package sem

import (
    "fmt"
    "strconv"
    "strings"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
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
            _ = res.Scope.Insert(&types.Object{Kind: types.ObjFunc, Name: fd.Name, Type: types.TInvalid})
            seen[fd.Name] = true
            funcs[fd.Name] = fd
            // Mutability analysis: default immutable, assignments require mut { }
            res.Diagnostics = append(res.Diagnostics, analyzeMutability(fd)...)
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
            res.Diagnostics = append(res.Diagnostics, analyzeWorkers(pd, funcs)...)
            res.Diagnostics = append(res.Diagnostics, analyzeIOPermissions(pd)...)
            res.Diagnostics = append(res.Diagnostics, analyzeEdges(pd)...)
            res.Diagnostics = append(res.Diagnostics, analyzeEdgeTypeSafety(pd, funcs)...)
        }
    }
    // Global type checks
    res.Diagnostics = append(res.Diagnostics, analyzeMapTypes(f)...)
    res.Diagnostics = append(res.Diagnostics, analyzeSetTypes(f)...)
    // Cross-pipeline cycle detection (unless cycle pragma present)
    res.Diagnostics = append(res.Diagnostics, analyzeCycles(f)...)
    return res
}

// analyzeEnum validates enum declarations: non-empty members, unique names,
// valid literal values (if provided), and disallow blank identifier members.
func analyzeEnum(ed astpkg.EnumDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if ed.Name == "" { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_NAME", Message: "enum must have a name"}) }
    if len(ed.Members) == 0 { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_EMPTY", Message: "enum has no members"}); return diags }
    seen := map[string]bool{}
    for _, m := range ed.Members {
        if m.Name == "_" { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_BLANK_MEMBER", Message: "enum member cannot be '_'"}) }
        if seen[m.Name] { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_DUP_MEMBER", Message: "duplicate enum member: "+m.Name}) }
        seen[m.Name] = true
        if m.Value != "" {
            if !(isIntLiteral(m.Value) || isStringLiteral(m.Value)) {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_VALUE_INVALID", Message: "enum member value must be integer or string literal: "+m.Name})
            }
        }
    }
    return diags
}

func isIntLiteral(s string) bool {
    if s == "" { return false }
    i := 0
    if s[0] == '-' { if len(s) == 1 { return false }; i = 1 }
    for ; i < len(s); i++ { if s[i] < '0' || s[i] > '9' { return false } }
    return true
}
func isStringLiteral(s string) bool {
    if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' { return true }
    return false
}

// analyzeStruct validates struct declarations: non-empty fields, unique names,
// non-blank field names, and presence of a type on each field.
func analyzeStruct(sd astpkg.StructDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if sd.Name == "" { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_NAME", Message: "struct must have a name"}) }
    if len(sd.Fields) == 0 { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_EMPTY", Message: "struct has no fields"}); return diags }
    seen := map[string]bool{}
    for _, f := range sd.Fields {
        if f.Name == "_" { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_BLANK_FIELD", Message: "struct field cannot be '_'"}) }
        if f.Name == "" { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_FIELD_NAME", Message: "struct field must have a name"}) }
        if seen[f.Name] { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_DUP_FIELD", Message: "duplicate struct field: "+f.Name}) }
        seen[f.Name] = true
        if f.Type.Name == "" && !f.Type.Ptr && !f.Type.Slice { // no recognizable type
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_FIELD_TYPE_INVALID", Message: "struct field missing or invalid type: "+f.Name})
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
                if k.Ptr || k.Slice { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be pointer or slice"}) }
                switch strings.ToLower(k.Name) {
                case "map", "set", "slice":
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be map/set/slice"})
                }
                if len(k.Args) > 0 { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be generic"}) }
            }
        }
        for _, a := range t.Args { walk(a) }
    }
    for _, d := range f.Decls {
        if sd, ok := d.(astpkg.StructDecl); ok {
            for _, fld := range sd.Fields { walk(fld.Type) }
        }
        if fd, ok := d.(astpkg.FuncDecl); ok {
            for _, p := range fd.Params { walk(p.Type) }
            for _, r := range fd.Result { walk(r) }
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
        for _, a := range t.Args { walk(a) }
    }
    for _, d := range f.Decls {
        if sd, ok := d.(astpkg.StructDecl); ok {
            for _, fld := range sd.Fields { walk(fld.Type) }
        }
        if fd, ok := d.(astpkg.FuncDecl); ok {
            for _, p := range fd.Params { walk(p.Type) }
            for _, r := range fd.Result { walk(r) }
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
    allowed := map[string]bool{"ingress":true, "transform":true, "fanout":true, "collect":true, "egress":true}
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
    if ingressCount > 1 { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_INGRESS", Message: fmt.Sprintf("pipeline %q has multiple ingress nodes", pd.Name)}) }
    if egressCount > 1 { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_EGRESS", Message: fmt.Sprintf("pipeline %q has multiple egress nodes", pd.Name)}) }

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
func analyzeWorkers(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    checkArgs := func(args []string) {
        for _, a := range args {
            name := a
            hasCall := false
            if i := strings.IndexRune(a, '('); i >= 0 { name = strings.TrimSpace(a[:i]); hasCall = true }
            // simple identifier extract: letters/_ followed by letters/digits/_
            if name == "" { continue }
            // skip placeholders like "cfg" or literals
            if name == "cfg" { continue }
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
            fd, ok := funcs[name]
            if !ok {
                // allow blank identifier '_' to pass worker ref check as sink
                if name == "_" { continue }
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

// analyzeMutability enforces that assignments occur only within explicit mut { } blocks.
// Implementation scans captured body tokens and tracks a parallel brace stack, marking
// frames opened by `mut {` as mutable. Any '=' token outside a mutable frame yields a diagnostic.
func analyzeMutability(fd astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if len(fd.Body) == 0 { return diags }
    mutStack := []bool{}
    mutDepth := 0
    expectMutLBrace := false
    for _, t := range fd.Body {
        switch t.Kind {
        case tok.KW_MUT:
            expectMutLBrace = true
        case tok.LBRACE:
            isMut := false
            if expectMutLBrace {
                isMut = true
                mutDepth++
                expectMutLBrace = false
            }
            mutStack = append(mutStack, isMut)
        case tok.RBRACE:
            if n := len(mutStack); n > 0 {
                wasMut := mutStack[n-1]
                mutStack = mutStack[:n-1]
                if wasMut && mutDepth > 0 { mutDepth-- }
            }
            expectMutLBrace = false
        case tok.ASSIGN:
            if mutDepth == 0 {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MUT_ASSIGN_OUTSIDE", Message: "assignment outside mut block is not allowed"})
            }
            expectMutLBrace = false
        default:
            // any other token cancels a pending mut if not followed by '{'
            // but we leave expectMutLBrace until a non-brace token? keep conservative
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
    for _, st := range pd.Steps { checkArgs(st.Args) }
    for _, st := range pd.ErrorSteps { checkArgs(st.Args) }
    return diags
}

// Minimal local spec structs to avoid cross-package dependency
type fifoSpec struct{ Min, Max int; BP, Type string }
type lifoSpec struct{ Min, Max int; BP, Type string }
type pipeSpec struct{ Name string; Min, Max int; BP, Type string }

// parseEdgeSpecFromArgs: copy of tolerant parser used in IR lowering (simplified)
func parseEdgeSpecFromArgs(args []string) (interface{}, bool) {
    for _, a := range args {
        s := strings.TrimSpace(a)
        if !strings.HasPrefix(s, "in=") { continue }
        v := strings.TrimPrefix(s, "in=")
        if strings.HasPrefix(v, "edge.FIFO(") && strings.HasSuffix(v, ")") {
            params := parseKVList(v[len("edge.FIFO(") : len(v)-1])
            var f fifoSpec
            for k, val := range params {
                switch k {
                case "minCapacity": f.Min = atoiSafe(val)
                case "maxCapacity": f.Max = atoiSafe(val)
                case "backpressure": f.BP = val
                case "type": f.Type = val
                }
            }
            return f, true
        }
        if strings.HasPrefix(v, "edge.LIFO(") && strings.HasSuffix(v, ")") {
            params := parseKVList(v[len("edge.LIFO(") : len(v)-1])
            var l lifoSpec
            for k, val := range params {
                switch k {
                case "minCapacity": l.Min = atoiSafe(val)
                case "maxCapacity": l.Max = atoiSafe(val)
                case "backpressure": l.BP = val
                case "type": l.Type = val
                }
            }
            return l, true
        }
        if strings.HasPrefix(v, "edge.Pipeline(") && strings.HasSuffix(v, ")") {
            params := parseKVList(v[len("edge.Pipeline(") : len(v)-1])
            var p pipeSpec
            for k, val := range params {
                switch k {
                case "name": p.Name = val
                case "minCapacity": p.Min = atoiSafe(val)
                case "maxCapacity": p.Max = atoiSafe(val)
                case "backpressure": p.BP = val
                case "type": p.Type = val
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
        if p == "" { continue }
        if eq := strings.IndexByte(p, '='); eq >= 0 {
            k := strings.TrimSpace(p[:eq])
            v := strings.TrimSpace(p[eq+1:])
            if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
                v = v[1:len(v)-1]
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
            if depth > 0 { depth-- }
        case ',':
            if depth == 0 {
                out = append(out, s[last:i])
                last = i+1
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
    if t.Ptr { b.WriteByte('*') }
    if t.Slice { b.WriteString("[]") }
    b.WriteString(t.Name)
    if len(t.Args) > 0 {
        b.WriteByte('<')
        for i, a := range t.Args {
            if i > 0 { b.WriteByte(',') }
            b.WriteString(typeRefToString(a))
        }
        b.WriteByte('>')
    }
    return b.String()
}

func isWorkerSignature(fd astpkg.FuncDecl) bool {
    // params: (Context, Event<T>, *State)
    if len(fd.Params) != 3 { return false }
    p1 := fd.Params[0].Type
    p2 := fd.Params[1].Type
    p3 := fd.Params[2].Type
    if !(p1.Name == "Context" && !p1.Ptr && !p1.Slice) { return false }
    if !(p2.Name == "Event" && len(p2.Args) == 1 && !p2.Ptr) { return false }
    if !(p3.Name == "State" && p3.Ptr) { return false }
    // results: exactly one of Event<U>, []Event<U>, Error<E>, Drop/Ack
    if len(fd.Result) != 1 { return false }
    r := fd.Result[0]
    switch {
    case r.Name == "Event" && len(r.Args) == 1 && !r.Slice:
        return true
    case r.Name == "Event" && len(r.Args) == 1 && r.Slice:
        return true
    case r.Name == "Error" && len(r.Args) == 1:
        return true
    case r.Name == "Drop" && len(r.Args) == 0:
        return true
    case r.Name == "Ack" && len(r.Args) == 0:
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
        if s == "" { return false }
        if strings.HasPrefix(s, "io=") { return true }
        if strings.Contains(s, "io.read(") || strings.Contains(s, "io.write(") { return true }
        return false
    }
    check := func(name string, args []string) {
        n := strings.ToLower(name)
        allowed := (n == "ingress" || n == "egress")
        if allowed { return }
        for _, a := range args {
            if isIOArg(a) {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_IO_PERMISSION", Message: "I/O operations are only allowed in ingress/egress nodes"})
                break
            }
        }
    }
    for _, st := range pd.Steps { check(st.Name, st.Args) }
    for _, st := range pd.ErrorSteps { check(st.Name, st.Args) }
    return diags
}
// analyzeEdgeTypeSafety validates that declared edge `type=` matches the
// upstream worker output payload type for each step.
func analyzeEdgeTypeSafety(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // helper to get worker result payload type string
    workerOut := func(name string) (string, bool) {
        fd, ok := funcs[name]
        if !ok { return "", false }
        if len(fd.Result) != 1 { return "", false }
        r := fd.Result[0]
        // Event<U> or []Event<U>
        if r.Name == "Event" && len(r.Args) == 1 { return typeRefToString(fd.Result[0].Args[0]), true }
        if r.Name == "Error" && len(r.Args) == 1 { return typeRefToString(fd.Result[0].Args[0]), true }
        return "", false
    }
    // Compare step i edge type to previous step workers' outputs
    for i := range pd.Steps {
        st := pd.Steps[i]
        spec, ok := parseEdgeSpecFromArgs(st.Args)
        if !ok { continue }
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
        if declared == "" { continue }
        // ensure previous step exists
        if i == 0 { continue }
        prev := pd.Steps[i-1]
        // Gather all worker outputs on previous step
        var outs []string
        for _, w := range prev.Workers {
            if t, ok := workerOut(w.Name); ok { outs = append(outs, t) }
        }
        // If there were no workers on previous step (e.g., Ingress), skip
        if len(outs) == 0 { continue }
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
        if strings.ToLower(d.Name) == "cycle" && strings.Contains(strings.ToLower(d.Payload), "allow") { allow = true; break }
    }
    if allow { return diags }
    // collect pipelines and edges
    names := []string{}
    idx := map[string]int{}
    for _, n := range f.Decls { if p, ok := n.(astpkg.PipelineDecl); ok { idx[p.Name] = len(names); names = append(names, p.Name) } }
    g := make([][]int, len(names))
    addEdge := func(from string, to string) {
        i, ok1 := idx[from]
        j, ok2 := idx[to]
        if ok1 && ok2 { g[i] = append(g[i], j) }
    }
    for _, n := range f.Decls {
        p, ok := n.(astpkg.PipelineDecl)
        if !ok { continue }
        scan := func(args []string) {
            if spec, ok := parseEdgeSpecFromArgs(args); ok {
                if v, ok2 := spec.(pipeSpec); ok2 {
                    if v.Name != "" { addEdge(p.Name, v.Name) }
                }
            }
        }
        for _, st := range p.Steps { scan(st.Args) }
        for _, st := range p.ErrorSteps { scan(st.Args) }
    }
    // detect cycles via DFS
    visited := make([]int, len(names)) // 0=unvisited,1=visiting,2=done
    var dfs func(int) bool
    dfs = func(u int) bool {
        visited[u] = 1
        for _, v := range g[u] {
            if visited[v] == 1 { return true }
            if visited[v] == 0 && dfs(v) { return true }
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
