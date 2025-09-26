package token

import "testing"

func TestKeywords_MapContainsCore(t *testing.T) {
    cases := map[string]Kind{
        "package": KW_PACKAGE,
        "import":  KW_IMPORT,
        "func":    KW_FUNC,
        "pipeline": KW_PIPELINE,
        "ingress": KW_INGRESS,
        "egress":  KW_EGRESS,
        "return":  KW_RETURN,
        "var":     KW_VAR,
    }
    for k, want := range cases {
        got, ok := Keywords[k]
        if !ok {
            t.Fatalf("keyword %q missing", k)
        }
        if got != want {
            t.Fatalf("keyword %q kind=%v want %v", k, got, want)
        }
    }
}

