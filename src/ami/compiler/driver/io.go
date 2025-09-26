package driver

import "os"

// mustReadFile returns source or empty string on error; build path already validated by caller.
func mustReadFile(path string) string {
    b, err := osReadFile(path)
    if err != nil {
        return ""
    }
    return string(b)
}

// osReadFile is indirection for testing/mocking.
var osReadFile = func(path string) ([]byte, error) { return os.ReadFile(path) }

