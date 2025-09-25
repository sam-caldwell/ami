package filepath

import stdpath "path"

// Join joins any number of path elements into a single path, separating with '/'.
func Join(elem ...string) string {
    if len(elem) == 0 { return "" }
    // Normalize each element to use '/'
    for i := range elem { elem[i] = normalizeSeparators(elem[i]) }
    return stdpath.Join(elem...)
}

