package os

// Minimal interface to avoid pulling io in public signature; concrete value is io.WriteCloser.
type ioWriteCloser interface{ Write([]byte) (int, error); Close() error }

