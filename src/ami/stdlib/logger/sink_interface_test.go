package logger

// Compile-time interface checks that sinks implement Sink.
var _ Sink = (*FileSink)(nil)
var _ Sink = (*StdoutSink)(nil)
var _ Sink = (*StderrSink)(nil)

