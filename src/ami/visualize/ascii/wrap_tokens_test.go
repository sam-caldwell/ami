package ascii

import "testing"

func TestWrapTokens_FilePair(t *testing.T) {
    got := wrapTokens([]string{"ab","cd","ef"}, 4)
    if got != "abcd\nef" { t.Fatalf("wrapTokens: %q", got) }
}

