package errorpipe

import "os"

// Default writes to os.Stderr. It is a convenience wrapper to mirror the
// default ErrorPipeline described in the spec (Ingress().Egress() → stderr).
func Default(code, message, file string, data map[string]any) error {
    return Write(os.Stderr, code, message, file, data)
}

