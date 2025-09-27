package edge

import "testing"

func TestMultiPath_Validate_Happy(t *testing.T) {
    mp := MultiPath{
        Attrs: map[string]any{"stable": true},
        Merge: []MergeAttr{{Name: "merge.Stable"}, {Name: "merge.Sort", Args: []any{"ts", "asc"}}},
    }
    if err := mp.Validate(); err != nil { t.Fatalf("validate: %v", err) }
    if mp.Kind() != KindMultiPath { t.Fatalf("kind") }
}

func TestMultiPath_Validate_Sad(t *testing.T) {
    mp := MultiPath{Merge: []MergeAttr{{Name: "merge.Unknown"}}}
    if err := mp.Validate(); err == nil { t.Fatalf("expected error for unknown merge attr") }
}

