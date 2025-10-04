package logger

// Basename pairing and compile-time interface check
var _ Sink = (*FileSink)(nil)

