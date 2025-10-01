package trigger

import "testing"

func TestFsNotify_NotImplemented(t *testing.T) {
    ch, stop, err := FsNotify("/tmp/does-not-matter", FsModify)
    if err == nil || ch != nil || stop != nil {
        t.Fatalf("expected not implemented error")
    }
}

