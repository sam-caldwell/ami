package main

import (
    "strings"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func attrsFromStep(st *ast.StepStmt) map[string]any {
    if st == nil { return nil }
    var bounded bool
    delivery := ""
    typ := ""
    multipath := ""
    for _, at := range st.Attrs {
        if (at.Name == "type" || at.Name == "Type") && len(at.Args) > 0 { typ = at.Args[0].Text }
        if len(at.Name) >= 6 && at.Name[:6] == "merge." {
            if at.Name == "merge.Buffer" {
                if len(at.Args) > 0 { if at.Args[0].Text != "0" && at.Args[0].Text != "" { bounded = true } }
                if len(at.Args) > 1 {
                    pol := at.Args[1].Text
                    switch pol {
                    case "dropOldest", "dropNewest": delivery = "bestEffort"
                    case "block": delivery = "atLeastOnce"
                    }
                }
            } else if at.Name == "merge.Shunt" {
                if len(at.Args) > 0 {
                    pol := at.Args[0].Text
                    switch pol {
                    case "newest": delivery = "shuntNewest"
                    case "oldest": delivery = "shuntOldest"
                    }
                }
            }
        }
        // Resolve configured edges: edge.FIFO / edge.LIFO
        if at.Name == "edge.FIFO" || at.Name == "edge.LIFO" {
            // parse kv
            kv := map[string]string{}
            for _, a := range at.Args { if eq := strings.IndexByte(a.Text, '='); eq > 0 { kv[strings.TrimSpace(a.Text[:eq])] = strings.TrimSpace(a.Text[eq+1:]) } }
            if t := kv["type"]; t != "" && typ == "" { typ = t }
            // capacity synonyms
            max := kv["max"]; if max == "" { max = kv["maxCapacity"] }
            if max != "" && max != "0" { bounded = true }
            switch kv["backpressure"] {
            case "dropOldest", "dropNewest": delivery = "bestEffort"
            case "block": delivery = "atLeastOnce"
            case "shuntNewest": delivery = "shuntNewest"
            case "shuntOldest": delivery = "shuntOldest"
            }
        }
        if at.Name == "edge.MultiPath" || at.Name == "MultiPath" {
            if len(at.Args) > 0 {
                var parts []string
                for _, a := range at.Args { if a.Text != "" { parts = append(parts, a.Text) } }
                if len(parts) > 0 { multipath = strings.Join(parts, "|") }
            }
        }
    }
    m := map[string]any{}
    if bounded { m["bounded"] = true }
    if delivery != "" { m["delivery"] = delivery }
    if typ != "" { m["type"] = typ }
    if multipath != "" { m["multipath"] = multipath }
    if len(m) == 0 { return nil }
    return m
}
