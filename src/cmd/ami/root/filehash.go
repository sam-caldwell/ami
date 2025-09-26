package root

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "os"
)

// fileSHA256 computes the sha256 and size of a file at path.
func fileSHA256(path string) (string, int64, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", 0, err
    }
    defer f.Close()
    h := sha256.New()
    n, err := io.Copy(h, f)
    if err != nil {
        return "", 0, err
    }
    return hex.EncodeToString(h.Sum(nil)), n, nil
}

