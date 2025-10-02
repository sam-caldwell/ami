package sem

import (
    "sort"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

func arityMismatchInTypesWithFields(et, at types.Type) (bool, []string, []int, []string, string, int, int) {
    switch ev := et.(type) {
    case types.Generic:
        av, ok := at.(types.Generic); if !ok { return false, nil, nil, nil, "", 0, 0 }
        if ev.Name != av.Name { return false, nil, nil, nil, "", 0, 0 }
        if len(ev.Args) != len(av.Args) { return true, []string{ev.Name}, []int{}, nil, ev.Name, len(ev.Args), len(av.Args) }
        for i := range ev.Args {
            if m, p, idx, fp, b, w, g := arityMismatchInTypesWithFields(ev.Args[i], av.Args[i]); m {
                return true, append([]string{ev.Name}, p...), append([]int{i}, idx...), fp, b, w, g
            }
        }
        return false, nil, nil, nil, "", 0, 0
    case types.Optional:
        av, ok := at.(types.Optional); if !ok { return false, nil, nil, nil, "", 0, 0 }
        return arityMismatchInTypesWithFields(ev.Inner, av.Inner)
    case types.Struct:
        av, ok := at.(types.Struct); if !ok { return false, nil, nil, nil, "", 0, 0 }
        // determine stable iteration order of common fields
        keys := make([]string, 0)
        for k := range ev.Fields { if _, ok := av.Fields[k]; ok { keys = append(keys, k) } }
        // stable by lexical order
        if len(keys) > 1 { sort.Strings(keys) }
        for _, k := range keys {
            einner := ev.Fields[k]
            ainner := av.Fields[k]
            if m, p, idx, fp, b, w, g := arityMismatchInTypesWithFields(einner, ainner); m {
                return true, p, idx, append([]string{k}, fp...), b, w, g
            }
        }
        return false, nil, nil, nil, "", 0, 0
    case types.Union:
        av, ok := at.(types.Union); if !ok { return false, nil, nil, nil, "", 0, 0 }
        // Try to find a pair of alts that reveals a mismatch; prefer matching by top-level constructor kind.
        for _, ealt := range ev.Alts {
            // prefer same kind
            for _, aalt := range av.Alts {
                switch ee := ealt.(type) {
                case types.Struct:
                    if aa, ok := aalt.(types.Struct); ok {
                        if m, p, idx, fp, b, w, g := arityMismatchInTypesWithFields(ee, aa); m { return true, p, idx, fp, b, w, g }
                    }
                case types.Optional:
                    if aa, ok := aalt.(types.Optional); ok {
                        if m, p, idx, fp, b, w, g := arityMismatchInTypesWithFields(ee, aa); m { return true, p, idx, fp, b, w, g }
                    }
                case types.Generic:
                    if aa, ok := aalt.(types.Generic); ok && ee.Name == aa.Name {
                        if m, p, idx, fp, b, w, g := arityMismatchInTypesWithFields(ee, aa); m { return true, p, idx, fp, b, w, g }
                    }
                default:
                    if m, p, idx, fp, b, w, g := arityMismatchInTypesWithFields(ee, aalt); m { return true, p, idx, fp, b, w, g }
                }
            }
        }
        return false, nil, nil, nil, "", 0, 0
    case types.Named:
        name := ev.Name
        if name == "any" || (len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z') { return false, nil, nil, nil, "", 0, 0 }
        return false, nil, nil, nil, "", 0, 0
    default:
        return false, nil, nil, nil, "", 0, 0
    }
}
