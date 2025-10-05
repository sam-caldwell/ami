package exec

import "testing"

func TestSanitizeWorkerSymbol(t *testing.T) {
    tests := []struct{ in, want string }{
        {"W", "prefixW"},
        {"my-worker", "prefixmy_worker"},
        {"9start", "prefixx_9start"},
        {"has space", "prefixhas_space"},
        {"many---dashes", "prefixmany_dashes"},
        {"", "prefix"},
    }
    for _, tt := range tests {
        got := SanitizeWorkerSymbol("prefix", tt.in)
        if got != tt.want {
            t.Fatalf("sanitize(%q) got %q want %q", tt.in, got, tt.want)
        }
    }
}

