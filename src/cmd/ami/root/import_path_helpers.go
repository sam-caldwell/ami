package root

import "strings"

func firstImportPathOffset(src, path string) int {
    // find first "import \"path\"" occurrence; crude search
    pat := "\"" + path + "\""
    return strings.Index(src, pat)
}

func importAliasOffset(src, alias, path string) int {
    // find "alias=\"path\"" or "alias \"path\"" patterns
    pat1 := alias + "=\"" + path + "\""
    if i := strings.Index(src, pat1); i >= 0 {
        return i
    }
    pat2 := alias + " \"" + path + "\""
    return strings.Index(src, pat2)
}

