package io

import (
    stdio "io"
    "errors"
    "os"
)

// ErrClosed is returned when an operation occurs on a closed handle.
var ErrClosed = errors.New("io.FHO: handle is closed")

// FHO is a file/stream handle abstraction for deterministic I/O semantics.
// Methods mirror Go's file operations but return (n int, err error) where applicable.
// After Close(), all operations fail with ErrClosed.
type FHO struct {
    f *os.File
}

// Open opens an existing file for reading (like os.Open).
func Open(name string) (*FHO, error) {
    f, err := os.Open(name)
    if err != nil { return nil, err }
    return &FHO{f: f}, nil
}

// Create creates or truncates a file for writing (like os.Create).
func Create(name string) (*FHO, error) {
    f, err := os.Create(name)
    if err != nil { return nil, err }
    return &FHO{f: f}, nil
}

// OpenFile is a general opener (like os.OpenFile) returning an FHO.
func OpenFile(name string, flag int, perm os.FileMode) (*FHO, error) {
    f, err := os.OpenFile(name, flag, perm)
    if err != nil { return nil, err }
    return &FHO{f: f}, nil
}

// Close closes the underlying handle; subsequent operations return ErrClosed.
func (h *FHO) Close() error {
    if h.f == nil { return nil }
    err := h.f.Close()
    h.f = nil
    return err
}

func (h *FHO) ensure() (*os.File, error) {
    if h == nil || h.f == nil { return nil, ErrClosed }
    return h.f, nil
}

// Read reads into p, returning bytes read and error.
func (h *FHO) Read(p []byte) (int, error) {
    f, err := h.ensure()
    if err != nil { return 0, err }
    return f.Read(p)
}

// ReadBytes is an alias for Read; included for API clarity.
func (h *FHO) ReadBytes(p []byte) (int, error) { return h.Read(p) }

// Write writes p to the file, returning bytes written and error.
func (h *FHO) Write(p []byte) (int, error) {
    f, err := h.ensure()
    if err != nil { return 0, err }
    return f.Write(p)
}

// WriteBytes is an alias for Write; included for API clarity.
func (h *FHO) WriteBytes(p []byte) (int, error) { return h.Write(p) }

// Seek sets the offset for the next Read/Write on file per POSIX whence.
func (h *FHO) Seek(offset int64, whence int) (int64, error) {
    f, err := h.ensure()
    if err != nil { return 0, err }
    return f.Seek(offset, whence)
}

// Pos returns the current file position for read/write operations.
func (h *FHO) Pos() (int64, error) { return h.Seek(0, stdio.SeekCurrent) }

// Length returns the current file length in bytes.
func (h *FHO) Length() (int64, error) {
    f, err := h.ensure()
    if err != nil { return 0, err }
    st, err := f.Stat()
    if err != nil { return 0, err }
    return st.Size(), nil
}

// Truncate truncates the file to size n.
func (h *FHO) Truncate(n int64) error {
    f, err := h.ensure()
    if err != nil { return err }
    return f.Truncate(n)
}

// Flush flushes file buffers to stable storage (fsync).
func (h *FHO) Flush() error {
    f, err := h.ensure()
    if err != nil { return err }
    return f.Sync()
}

