package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestDebug_Tokenize_ReturnsTwice(t *testing.T) {
    src := "package app\nfunc H(){ return slice<int>{1}; return slice<string>{\"x\"} }"
    f := &source.File{Name: "ri_tok.ami", Content: src}
    s := scanner.New(f)
    count := 0
    for {
        t := s.Next()
        if t.Kind == token.EOF { break }
        if t.Kind == token.KwReturn { count++ }
    }
    if count != 2 { t.Fatalf("expected 2 return tokens, got %d", count) }
}

