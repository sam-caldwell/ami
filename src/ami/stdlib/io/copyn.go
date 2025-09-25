package io

import stdio "io"

// CopyN copies n bytes (or until EOF) from src to dst using deterministic, blocking semantics.
// It returns the number of bytes copied and the first error encountered while copying.
func CopyN(dst stdio.Writer, src stdio.Reader, n int64) (int64, error) { return stdio.CopyN(dst, src, n) }

