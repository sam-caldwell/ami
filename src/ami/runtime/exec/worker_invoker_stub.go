//go:build !cgo

package exec

// NewDLSOInvoker is unavailable without cgo; return nil for graceful fallback.
func NewDLSOInvoker(libPath, prefix string) WorkerInvoker { return nil }

