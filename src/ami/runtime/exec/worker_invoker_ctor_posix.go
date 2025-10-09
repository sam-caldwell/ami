//go:build cgo && (linux || darwin)

package exec

// NewDLSOInvoker creates an invoker for the given shared library. When libPath is empty,
// symbols are resolved from the current process (RTLD_DEFAULT).
func NewDLSOInvoker(libPath, prefix string) *DLSOInvoker { return &DLSOInvoker{libPath: libPath, prefix: prefix} }

