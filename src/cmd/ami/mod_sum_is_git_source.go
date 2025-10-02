package main

func isGitSource(s string) bool {
    return len(s) > 0 && (hasPrefix(s, "git+ssh://") || hasPrefix(s, "file+git://"))
}

