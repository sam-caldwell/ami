package mod

import "testing"

func TestToSSHURL_Normalization(t *testing.T) {
    cases := map[string]string{
        "github.com/org/repo":           "ssh://git@github.com/org/repo.git",
        "github.com/org/repo.git":       "ssh://git@github.com/org/repo.git",
        "ssh://git@host/x/y.git":        "ssh://git@host/x/y.git",
        "git@host:x/y.git":              "git@host:x/y.git", // passthrough non-ssh scheme is kept as-is
    }
    for in, want := range cases {
        got := toSSHURL(in)
        if got != want {
            t.Fatalf("toSSHURL(%q) = %q; want %q", in, got, want)
        }
    }
}

