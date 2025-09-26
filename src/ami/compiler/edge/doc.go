// Package edge contains declarative, compiler-generated edge specifications.
//
// In AMI source, authors write constructs like:
//   in=edge.FIFO(minCapacity=10, maxCapacity=20, backpressure=block, type=T)
//   in=edge.LIFO(...)
//   in=edge.Pipeline(name=someIngressName, ...)
// These are not runtime function calls. The compiler recognizes these node
// arguments and generates efficient queue/linkage implementations in codegen.
//
// This package provides stub types used by the compiler/IR phases to carry
// configuration. The concrete queue implementations are emitted by codegen and
// tuned for high performance (lock-free or low-contention ring buffers, cache
// friendliness, pooling, and zero-copy where safe).
package edge

