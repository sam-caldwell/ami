package root

import "os"

// hasAnyWritePermission returns true if any write bit is set on dir.
func hasAnyWritePermission(dir string) bool {
    fi, err := os.Stat(dir)
    if err != nil {
        return false
    }
    // coarse check: any of user/group/other write bits
    return fi.Mode().Perm()&0o222 != 0
}

