package sem

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    cty "github.com/sam-caldwell/ami/src/ami/compiler/types"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeEventTypeFlow performs a conservative type-flow check across pipeline edges.
// If both the upstream and downstream steps declare a `type(...)` attribute with a
// single string or ident argument, and they differ, emit E_EVENT_TYPE_FLOW at the
// downstream step position.
func AnalyzeEventTypeFlow(f *ast.File) []diag.Record {
    return AnalyzeEventTypeFlowInContext(f, nil)
}

// AnalyzeEventTypeFlowInContext extends AnalyzeEventTypeFlow by consulting a
// package-level egress type registry for cross-pipeline propagation at
// edge.Pipeline(name=...). When egressType is nil, it behaves like the local
// analysis.
func AnalyzeEventTypeFlowInContext(f *ast.File, egressType map[string]string) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        // collect step types and positions
        stepType := map[string]string{}
        stepPos := map[string]diag.Position{}
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                stepPos[st.Name] = diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
                for _, at := range st.Attrs {
                    if at.Name == "type" || at.Name == "Type" {
                        if len(at.Args) > 0 && at.Args[0].Text != "" {
                            stepType[st.Name] = at.Args[0].Text
                        }
                    } else if at.Name == "edge.Pipeline" && egressType != nil {
                        // propagate type from pipeline registry when available
                        if len(at.Args) > 0 {
                            for _, a := range at.Args {
                                kv := a.Text
                                if eq := strings.IndexByte(kv, '='); eq > 0 {
                                    k := strings.TrimSpace(kv[:eq])
                                    v := strings.TrimSpace(kv[eq+1:])
                                    if k == "name" {
                                        if t := egressType[v]; t != "" && stepType[st.Name] == "" {
                                            stepType[st.Name] = t
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
        // build edge list and indegree counts
        type edge struct{ from, to string }
        var edges []edge
        // compare across edges
        for _, s := range pd.Stmts {
            if e, ok := s.(*ast.EdgeStmt); ok {
                edges = append(edges, edge{from: e.From, to: e.To})
            }
        }
        // simple iterative propagation: forward then backward passes
        // forward: if all preds have same non-empty type, assign to node
        for pass := 0; pass < 2; pass++ {
            changed := false
            for _, e := range edges {
                // forward check for target if unset
                if stepType[e.to] == "" {
                    // gather preds of e.to
                    t := ""
                    consistent := true
                    for _, ee := range edges {
                        if ee.to == e.to {
                            pt := stepType[ee.from]
                            if pt == "" { consistent = false; break }
                            if t == "" { t = pt } else if t != pt { consistent = false; break }
                        }
                    }
                    if consistent && t != "" { stepType[e.to] = t; changed = true }
                }
                // backward: if source unset and all succs have same type
                if stepType[e.from] == "" {
                    t := ""
                    consistent := true
                    for _, ee := range edges {
                        if ee.from == e.from {
                            pt := stepType[ee.to]
                            if pt == "" { consistent = false; break }
                            if t == "" { t = pt } else if t != pt { consistent = false; break }
                        }
                    }
                    if consistent && t != "" { stepType[e.from] = t; changed = true }
                }
            }
            if !changed { break }
        }
        // emit mismatches where both sides have explicit/derived types
        for _, e := range edges {
            tFrom := stepType[e.from]
            tTo := stepType[e.to]
            if tFrom != "" && tTo != "" && !compatibleEventTypes(tTo, tFrom) {
                pos := stepPos[e.to]
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EVENT_TYPE_FLOW", Message: "event type mismatch across edge", Pos: &pos, Data: map[string]any{"from": e.from, "to": e.to, "fromType": tFrom, "toType": tTo}})
            }
        }
        // MultiPath inputs: if Collect has >1 typed upstream with differing types, flag mismatch at Collect
        // gather incoming types per node
        inTypes := map[string]map[string]struct{}{}
        for _, e := range edges {
            if stepType[e.from] != "" {
                if inTypes[e.to] == nil { inTypes[e.to] = map[string]struct{}{} }
                inTypes[e.to][stepType[e.from]] = struct{}{}
            }
        }
        for node, set := range inTypes {
            if len(set) > 1 {
                pos := stepPos[node]
                // collect list of types for diagnostics
                types := make([]string, 0, len(set))
                for k := range set { types = append(types, k) }
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EVENT_TYPE_FLOW", Message: "event types differ across multiple upstreams", Pos: &pos, Data: map[string]any{"node": node, "types": types}})
            }
        }
    }
    return out
}

// compatibleEventTypes returns true when downstream 'expected' event type is compatible with
// upstream 'actual'. It first applies simple text compatibility, then augments for Event<Union<...>>
// where membership of the actual payload type within the expected union is considered compatible.
func compatibleEventTypes(expected, actual string) bool {
    // fast path using existing textual rules
    if typesCompatible(expected, actual) { return true }
    // structural check for Event<...>
    if !strings.HasPrefix(expected, "Event<") || !strings.HasPrefix(actual, "Event<") { return false }
    et, err1 := cty.Parse(expected)
    at, err2 := cty.Parse(actual)
    if err1 != nil || err2 != nil { return false }
    // Unwrap Event payloads
    eg, ok1 := et.(cty.Generic)
    ag, ok2 := at.(cty.Generic)
    if !ok1 || !ok2 || eg.Name != "Event" || ag.Name != "Event" || len(eg.Args) != 1 || len(ag.Args) != 1 { return false }
    exp := eg.Args[0]
    act := ag.Args[0]
    // Direct structural equality
    if cty.Equal(exp, act) { return true }
    // If expected is Optional<X>, allow actual X or Optional<X> (including union membership beneath Optional).
    if eo, ok := exp.(cty.Optional); ok {
        inner := eo.Inner
        // If actual is Optional<Y>, compare Y against inner; else compare actual directly against inner.
        if ao, ok := act.(cty.Optional); ok {
            return payloadWithin(inner, ao.Inner)
        }
        return payloadWithin(inner, act)
    }
    // Otherwise check if expected is a Union and actual is a member.
    return payloadWithin(exp, act)
}

// payloadWithin reports whether actual equals expected or, when expected is a Union, whether
// actual is one of the union alternatives.
func payloadWithin(expected, actual cty.Type) bool {
    // Exact structural match
    if cty.Equal(expected, actual) { return true }

    // Optional covariance: Optional<X> accepts X or Optional<Y> when Y ∈ X (recursively).
    if eo, ok := expected.(cty.Optional); ok {
        // If actual is Optional<Y>, compare Y to X; else compare actual to X directly.
        if ao, ok := actual.(cty.Optional); ok { return payloadWithin(eo.Inner, ao.Inner) }
        return payloadWithin(eo.Inner, actual)
    }

    // Union membership: expected is Union, actual must be a member.
    if eu, ok := expected.(cty.Union); ok {
        for _, alt := range eu.Alts { if payloadWithin(alt, actual) { return true } }
        return false
    }

    // Struct width subtyping: expected struct fields must be present and compatible in actual.
    if es, ok := expected.(cty.Struct); ok {
        if as, ok := actual.(cty.Struct); ok {
            for k, et := range es.Fields {
                at, ok := as.Fields[k]
                if !ok { return false }
                // Allow Optional wrappers on actual fields when expected is non-Optional, by unwrapping once here.
                if ao, ok := at.(cty.Optional); ok { at = ao.Inner }
                if !payloadWithin(et, at) { return false }
            }
            return true
        }
        return false
    }

    // Container variance: same generic name with compatible type arguments.
    // Supported forms via Parse: slice<T>, set<T>, map<K,V>, Owned<T>, Error<E>, etc.
    if eg, ok := expected.(cty.Generic); ok {
        if ag, ok := actual.(cty.Generic); ok && eg.Name == ag.Name && len(eg.Args) == len(ag.Args) {
            for i := range eg.Args { if !payloadWithin(eg.Args[i], ag.Args[i]) { return false } }
            return true
        }
    }

    // No match by structure.
    return false
}
