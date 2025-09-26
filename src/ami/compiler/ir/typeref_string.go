package ir

import (
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// typeRefToString renders an AST TypeRef (without positions) into a concise string.
func typeRefToString(t astpkg.TypeRef) string {
    var b strings.Builder
    if t.Ptr {
        b.WriteByte('*')
    }
    if t.Slice {
        b.WriteString("[]")
    }
    b.WriteString(t.Name)
    if len(t.Args) > 0 {
        b.WriteByte('<')
        for i, a := range t.Args {
            if i > 0 {
                b.WriteByte(',')
            }
            b.WriteString(typeRefToString(a))
        }
        b.WriteByte('>')
    }
    return b.String()
}

