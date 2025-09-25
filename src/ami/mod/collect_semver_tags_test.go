package mod

import (
    "testing"
    "github.com/go-git/go-git/v5/plumbing"
)

func TestCollectSemverTags(t *testing.T) {
    refs := []*plumbing.Reference{
        plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/main"), plumbing.ZeroHash),
        plumbing.NewHashReference(plumbing.ReferenceName("refs/tags/v1.0.0"), plumbing.ZeroHash),
        plumbing.NewHashReference(plumbing.ReferenceName("refs/tags/not-a-ver"), plumbing.ZeroHash),
        plumbing.NewHashReference(plumbing.ReferenceName("refs/tags/v1.0.1"), plumbing.ZeroHash),
    }
    tags := collectSemverTags(refs)
    if len(tags) != 2 { t.Fatalf("expected 2 semver tags; got %v", tags) }
    sortSemver(tags)
    if tags[0] != "v1.0.0" || tags[1] != "v1.0.1" { t.Fatalf("unexpected tags: %v", tags) }
}

