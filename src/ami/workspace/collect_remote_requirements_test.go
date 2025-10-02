package workspace

import "testing"

func TestCollectRemoteRequirements_Empty(t *testing.T) {
    ws := Workspace{}
    r, errs := CollectRemoteRequirements(&ws)
    if len(r) != 0 || len(errs) != 0 { t.Fatalf("unexpected: %v %v", r, errs) }
}

