package main

import "testing"

// compile-time check only
type _ioWriterImpl struct{}
func (_ioWriterImpl) Write(b []byte) (int, error) { return len(b), nil }

func Test_ioWriter_iface(t *testing.T) { var _ ioWriter = _ioWriterImpl{} }

