package main

import "regexp"

// validPkgName returns true when name conforms to allowed styles: lowercase, camelCase, or PascalCase; underscores disallowed.
func validPkgName(name string) bool {
    if name == "" { return false }
    if containsUnderscore(name) { return false }
    // lowercase
    if regexp.MustCompile(`^[a-z][a-z0-9]*$`).MatchString(name) { return true }
    // camelCase
    if regexp.MustCompile(`^[a-z][A-Za-z0-9]*$`).MatchString(name) { return true }
    // PascalCase
    if regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`).MatchString(name) { return true }
    return false
}

