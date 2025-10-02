package sem

import (
    "regexp"
    "strings"
)

var durRe = regexp.MustCompile(`^\d+(ms|s|m|h)$`)

func validPositiveDuration(s string) bool {
    s = strings.TrimSpace(s)
    if s == "" { return false }
    return durRe.MatchString(s)
}
