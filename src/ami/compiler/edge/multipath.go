package edge

import "fmt"

// MultiPath configures Collect-node merge behavior via merge.* attributes and/or
// simple k=v pairs represented through Attrs. Validation is minimal; deeper
// semantic checks occur in the sem package.
type MultiPath struct {
    Attrs map[string]any
    Merge []MergeAttr
}

func (m MultiPath) Kind() Kind { return KindMultiPath }

var allowedMerge = map[string]struct{}{
    "merge.Sort": {},
    "merge.Stable": {},
    "merge.Key": {},
    "merge.Dedup": {},
    "merge.Window": {},
    "merge.Watermark": {},
    "merge.Timeout": {},
    "merge.Buffer": {},
    "merge.PartitionBy": {},
}

func (m MultiPath) Validate() error {
    for k := range m.Attrs {
        if k == "" {
            return fmt.Errorf("attribute key cannot be empty")
        }
    }
    for _, a := range m.Merge {
        if a.Name == "" { return fmt.Errorf("merge attribute name cannot be empty") }
        if _, ok := allowedMerge[a.Name]; !ok {
            return fmt.Errorf("unknown merge attribute: %q", a.Name)
        }
    }
    return nil
}
